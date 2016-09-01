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

type errorReader struct{}

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

	defer db.Close()

	Convey("Routes", t, func() {
		db.Clean()

		wr := httptest.NewRecorder()
		m := mux.NewRouter()

		Convey("GetTaskInfo", func() {
			r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)
			m.HandleFunc("/task/{taskId}", h.GetTaskInfo(&config.Config{}))

			Convey("JSON request", func() {
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

			Convey("text/html request", func() {
				r.Header.Add("Accept", "text/html")

				Convey("Not Found", func() {
					id := "eremetic-task.9876"
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

					b, _ := ioutil.ReadAll(wr.Body)
					body := string(b)
					So(wr.Code, ShouldEqual, http.StatusNotFound)
					So(body, ShouldContainSubstring, "<title>404 Not Found | Eremetic</title>")
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

					b, _ := ioutil.ReadAll(wr.Body)
					body := string(b)
					So(wr.Code, ShouldEqual, http.StatusOK)
					lookup := fmt.Sprintf("<body data-task=\"%s\">", id)
					So(body, ShouldContainSubstring, lookup)
				})
			})
		})

		Convey("AddTask", func() {
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

		Convey("Sandbox Paths", func() {
			Convey("Get Files from sandbox", func() {
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
		})

		Convey("Version", func() {
			r, _ := http.NewRequest("GET", "/version", nil)
			m.HandleFunc("/version", h.Version(&config.Config{Version: "test"}))
			m.ServeHTTP(wr, r)

			body, _ := ioutil.ReadAll(wr.Body)
			So(string(body), ShouldEqual, "\"test\"\n")
		})

		Convey("Index", func() {
			Convey("Renders nothing for json", func() {
				r, _ := http.NewRequest("GET", "/", nil)
				m.HandleFunc("/", h.IndexHandler(&config.Config{Version: "test"}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusNoContent)
			})

			Convey("Renders the html template for html requests", func() {
				r, _ := http.NewRequest("GET", "/", nil)
				r.Header.Add("Accept", "text/html")
				m.HandleFunc("/", h.IndexHandler(&config.Config{Version: "test"}))
				m.ServeHTTP(wr, r)

				b, _ := ioutil.ReadAll(wr.Body)
				body := string(b)
				So(wr.Code, ShouldEqual, http.StatusOK)
				So(body, ShouldContainSubstring, "<html>")
				So(body, ShouldContainSubstring, "<div id='eremetic-version'>test</div>")
			})
		})

	})
}
