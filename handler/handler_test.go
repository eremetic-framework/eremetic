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
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/mocks"
	"github.com/klarna/eremetic/types"
	"github.com/klarna/eremetic/version"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandling(t *testing.T) {
	scheduler := &mocks.Scheduler{}
	status := []types.Status{
		types.Status{
			Status: types.TaskState_TASK_RUNNING,
			Time:   time.Now().Unix(),
		},
	}

	dir := os.TempDir()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/eremetic_test.db", dir))
	if err != nil {
		t.Fail()
	}
	h := Create(scheduler, db)

	defer db.Close()

	Convey("Routes", t, func() {
		db.Clean()

		id := "eremetic-task.1234"

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

		wr := httptest.NewRecorder()
		m := mux.NewRouter()
		r, _ := http.NewRequest("GET", fmt.Sprintf("/task/%s", id), nil)

		Convey("GetTaskInfo", func() {
			m.HandleFunc("/task/{taskId}", h.GetTaskInfo(&config.Config{}))

			Convey("JSON request", func() {
				Convey("Not Found", func() {
					id = "eremetic-task.5678"
					r.URL, _ = url.Parse(fmt.Sprintf("/task/%s", id))

					db.PutTask(&task)
					m.ServeHTTP(wr, r)

					So(wr.Code, ShouldEqual, http.StatusNotFound)
				})

				Convey("Found", func() {
					db.PutTask(&task)
					m.ServeHTTP(wr, r)

					So(wr.Code, ShouldEqual, http.StatusOK)
				})

				Convey("Task with MaskedEnv gets masked", func() {
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
					id = "eremetic-task.5678"
					r.URL, _ = url.Parse(fmt.Sprintf("/task/%s", id))

					db.PutTask(&task)
					m.ServeHTTP(wr, r)

					b, _ := ioutil.ReadAll(wr.Body)
					body := string(b)

					So(wr.Code, ShouldEqual, http.StatusNotFound)
					So(body, ShouldContainSubstring, "<title>404 Not Found | Eremetic</title>")
				})

				Convey("Found", func() {
					db.PutTask(&task)
					m.ServeHTTP(wr, r)

					b, _ := ioutil.ReadAll(wr.Body)
					body := string(b)
					So(wr.Code, ShouldEqual, http.StatusOK)
					lookup := fmt.Sprintf("<body data-task=\"%s\">", id)
					So(body, ShouldContainSubstring, lookup)
					status := "<div class=\"ui task_running label\">"
					So(body, ShouldContainSubstring, status)
				})
			})
		})

		Convey("AddTask", func() {
			data := []byte(`{"task_mem":22.0, "docker_image": "busybox", "command": "echo hello", "task_cpus":0.5, "tasks_to_launch": 1}`)
			r, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(data))
			r.Host = "localhost"

			Convey("It should respond with a location header", func() {
				handler := h.AddTask()
				handler(wr, r)

				location := wr.HeaderMap["Location"][0]
				So(location, ShouldStartWith, "http://localhost/task/eremetic-task.")
				So(wr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("Failed to schedule", func() {
				err := errors.New("A random error")
				scheduler.NextError = &err

				handler := h.AddTask()
				handler(wr, r)

				So(wr.Code, ShouldEqual, 500)
			})

			Convey("Error on bad input stream", func() {
				r.Body = ioutil.NopCloser(&mocks.ErrorReader{})

				handler := h.AddTask()
				handler(wr, r)

				So(wr.Code, ShouldEqual, 422)
			})

			Convey("Error on malformed json", func() {
				data = []byte(`{"key:123}`)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

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
			version.Version = "test"
			r, _ := http.NewRequest("GET", "/version", nil)
			m.HandleFunc("/version", h.Version(&config.Config{}))
			m.ServeHTTP(wr, r)

			body, _ := ioutil.ReadAll(wr.Body)
			So(string(body), ShouldEqual, "\"test\"\n")
		})

		Convey("Index", func() {
			r, _ := http.NewRequest("GET", "/", nil)

			Convey("Renders nothing for json", func() {
				m.HandleFunc("/", h.IndexHandler(&config.Config{}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusNoContent)
			})

			Convey("Renders the html template for html requests", func() {
				r.Header.Add("Accept", "text/html")
				m.HandleFunc("/", h.IndexHandler(&config.Config{}))
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
