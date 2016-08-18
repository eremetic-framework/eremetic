package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/types"
	. "github.com/smartystreets/goconvey/convey"
)

type mockError struct {
	message string
}

func (m mockError) Error() string {
	return m.message
}

type mockScheduler struct {
	nextError *error
}

func (s *mockScheduler) ScheduleTask(request types.Request) (string, error) {
	if err := s.nextError; err != nil {
		s.nextError = nil
		return "", *err
	}
	return "eremetic-task.mock", nil
}

type errorReader struct {
}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("oh no")
}

func TestHandling(t *testing.T) {
	scheduler := &mockScheduler{}
	status := []types.Status{
		types.Status{
			Status: types.TaskState_TASK_RUNNING,
			Time:   time.Now().Unix(),
		},
	}

	dir, _ := os.Getwd()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil {
		t.Fail()
	}
	h := Create(scheduler, db)

	db.Clean()
	defer db.Close()

	Convey("writeJSON", t, func() {
		Convey("Should respond with a JSON and the appropriate status code", func() {
			var wr = httptest.NewRecorder()

			writeJSON(200, "foo", wr)
			contentType := wr.HeaderMap["Content-Type"][0]
			So(contentType, ShouldEqual, "application/json; charset=UTF-8")
			So(wr.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("HandleError", t, func() {
		wr := httptest.NewRecorder()

		Convey("It should return an error status code", func() {
			var err = mockError{
				message: "Error",
			}

			handleError(err, wr, "A test error")

			So(wr.Code, ShouldEqual, 422)
			So(strings.TrimSpace(wr.Body.String()), ShouldEqual, "{\"error\":\"Error\",\"message\":\"A test error\"}")
		})
	})

	Convey("GetTaskInfo", t, func() {
		wr := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)
		m := mux.NewRouter()
		m.HandleFunc("/task/{taskId}", h.GetTaskInfo(&config.Config{}))

		Convey("Not Found", func() {
			id := "eremetic-task.5678"
			task := types.EremeticTask{
				TaskCPUs: 0.2,
				TaskMem:  0.5,
				Command:  "test",
				Image:    "test",
				Status:   status,
				ID:       id,
			}
			db.PutTask(&task)
			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Found", func() {
			id := "eremetic-task.1234"
			task := types.EremeticTask{
				TaskCPUs: 0.2,
				TaskMem:  0.5,
				Command:  "test",
				Image:    "test",
				Status:   status,
				ID:       id,
			}
			db.PutTask(&task)
			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Task with MaskedEnv gets masked", func() {
			r, _ := http.NewRequest("GET", "/task/eremetic-task.987", nil)
			id := "eremetic-task.987"
			maskedEnv := make(map[string]string)
			maskedEnv["foo"] = "bar"
			task := types.EremeticTask{
				TaskCPUs:          0.2,
				TaskMem:           0.5,
				Command:           "test",
				Image:             "test",
				Status:            status,
				ID:                id,
				MaskedEnvironment: maskedEnv,
			}

			db.PutTask(&task)
			m.ServeHTTP(wr, r)

			var retrievedTask types.EremeticTask
			body, _ := ioutil.ReadAll(io.LimitReader(wr.Body, 1048576))

			json.Unmarshal(body, &retrievedTask)
			So(retrievedTask.MaskedEnvironment, ShouldContainKey, "foo")
			So(retrievedTask.MaskedEnvironment["foo"], ShouldNotEqual, "bar")
			So(retrievedTask.MaskedEnvironment["foo"], ShouldEqual, "*******")

		})
	})

	Convey("AddTask", t, func() {
		wr := httptest.NewRecorder()

		Convey("It should respond with a location header", func() {
			data := []byte(`{"task_mem":22.0, "docker_image": "busybox", "command": "echo hello", "task_cpus":0.5, "tasks_to_launch": 1}`)
			r, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(data))
			r.Host = "localhost"

			handler := h.AddTask()
			handler(wr, r)

			location := wr.HeaderMap["Location"][0]
			So(location, ShouldStartWith, "http://localhost/task/eremetic-task.")
			So(wr.Code, ShouldEqual, http.StatusAccepted)
		})

		Convey("Failed to schedule", func() {
			data := []byte(`{"task_mem":22.0, "docker_image": "busybox", "command": "echo hello", "task_cpus":0.5, "tasks_to_launch": 1}`)
			r, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(data))
			r.Host = "localhost"
			err := errors.New("A random error")
			scheduler.nextError = &err

			handler := h.AddTask()
			handler(wr, r)

			So(wr.Code, ShouldEqual, 500)
		})

		Convey("Error on bad input stream", func() {
			r, _ := http.NewRequest("POST", "/task", &errorReader{})
			r.Host = "localhost"

			handler := h.AddTask()
			handler(wr, r)

			So(wr.Code, ShouldEqual, 422)
		})

		Convey("Error on malformed json", func() {
			data := []byte(`{"key:123}`)
			r, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(data))
			r.Host = "localhost"

			handler := h.AddTask()
			handler(wr, r)

			So(wr.Code, ShouldEqual, 422)
		})
	})

	Convey("Get Files from sandbox", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintf(w, "mocked")
		}))
		defer s.Close()

		addr := strings.Split(s.Listener.Addr().String(), ":")
		ip := addr[0]
		port, _ := strconv.ParseInt(addr[1], 10, 32)
		id := "eremetic-task.1234"

		task := types.EremeticTask{
			TaskCPUs:    0.2,
			TaskMem:     0.5,
			Command:     "test",
			Image:       "test",
			Status:      status,
			ID:          id,
			SandboxPath: "/tmp",
			AgentIP:     ip,
			AgentPort:   int32(port),
		}
		db.PutTask(&task)
		wr := httptest.NewRecorder()
		m := mux.NewRouter()

		Convey("stdout", func() {
			r, _ := http.NewRequest("GET", "/task/eremetic-task.1234/stdout", nil)
			m.HandleFunc("/task/{taskId}/stdout", h.GetFromSandbox("stdout"))
			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusOK)
			So(wr.Header().Get("Content-Type"), ShouldEqual, "text/plain; charset=UTF-8")

			body, _ := ioutil.ReadAll(wr.Body)
			So(string(body), ShouldEqual, "mocked")
		})

		Convey("stderr", func() {
			r, _ := http.NewRequest("GET", "/task/eremetic-task.1234/stderr", nil)
			m.HandleFunc("/task/{taskId}/stderr", h.GetFromSandbox("stderr"))

			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusOK)
			So(wr.Header().Get("Content-Type"), ShouldEqual, "text/plain; charset=UTF-8")

			body, _ := ioutil.ReadAll(wr.Body)
			So(string(body), ShouldEqual, "mocked")
		})

		Convey("No Sandbox path", func() {
			r, _ := http.NewRequest("GET", "/task/eremetic-task.1234/stdout", nil)
			m.HandleFunc("/task/{taskId}/stdout", h.GetFromSandbox("stdout"))

			task := types.EremeticTask{
				TaskCPUs:    0.2,
				TaskMem:     0.5,
				Command:     "test",
				Image:       "test",
				Status:      status,
				ID:          id,
				SandboxPath: "",
				AgentIP:     ip,
				AgentPort:   int32(port),
			}
			db.PutTask(&task)

			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusNoContent)
		})

	})

	Convey("renderHTML", t, func() {
		id := "eremetic-task.1234"

		task := types.EremeticTask{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "test",
			Image:    "test",
			Status:   status,
			ID:       id,
		}

		wr := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)

		renderHTML(wr, r, task, id, &config.Config{Version: "test"})

		body, _ := ioutil.ReadAll(wr.Body)
		So(body, ShouldNotBeEmpty)
		So(string(body), ShouldContainSubstring, "html")
	})

	Convey("makeMap", t, func() {
		task := types.EremeticTask{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "test",
			Image:    "test",
			Status:   status,
			ID:       "eremetic-task.1234",
		}

		data := makeMap(task)
		So(data, ShouldContainKey, "CPU")
		So(data, ShouldContainKey, "Memory")
		So(data, ShouldContainKey, "Status")
		So(data, ShouldContainKey, "ContainerImage")
		So(data, ShouldContainKey, "Command")
		So(data, ShouldContainKey, "TaskID")
	})
}
