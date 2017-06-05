package server

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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/config"
	"github.com/eremetic-framework/eremetic/mock"
	"github.com/eremetic-framework/eremetic/version"
)

func TestHandling(t *testing.T) {
	scheduler := &mock.ErrScheduler{}
	status := []eremetic.Status{
		eremetic.Status{
			Status: eremetic.TaskRunning,
			Time:   time.Now().Unix(),
		},
	}

	db := eremetic.NewDefaultTaskDB()

	h := NewHandler(scheduler, db)

	defer db.Close()

	Convey("Routes", t, func() {

		id := "eremetic-task.1234"

		maskedEnv := make(map[string]string)
		maskedEnv["foo"] = "bar"
		task := eremetic.Task{
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

					var retrievedTask eremetic.Task
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
				handler := h.AddTask(&config.Config{})
				handler(wr, r)

				location := wr.HeaderMap["Location"][0]
				So(location, ShouldStartWith, "http://localhost/task/eremetic-task.")
				So(wr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("It should respond with a location per URL prefix header", func() {
				conf := config.Config{}
				conf.URLPrefix = "/service/eremetic"
				handler := h.AddTask(&conf)
				handler(wr, r)

				location := wr.HeaderMap["Location"][0]
				So(location, ShouldStartWith, "http://localhost/service/eremetic/task/eremetic-task.")
				So(wr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("Failed to schedule", func() {
				err := errors.New("A random error")
				scheduler.NextError = &err

				handler := h.AddTask(&config.Config{})
				handler(wr, r)

				So(wr.Code, ShouldEqual, 500)
			})

			Convey("Error on bad input stream", func() {
				r.Body = ioutil.NopCloser(&mock.ErrorReader{})

				handler := h.AddTask(&config.Config{})
				handler(wr, r)

				So(wr.Code, ShouldEqual, 422)
			})

			Convey("Error on malformed json", func() {
				data = []byte(`{"key:123}`)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

				handler := h.AddTask(&config.Config{})
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

				task := eremetic.Task{
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

					task := eremetic.Task{
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

		Convey("Delete task", func() {
			r, _ := http.NewRequest("DELETE", fmt.Sprintf("/task/%s", id), nil)
			m.HandleFunc("/task/{taskId}", h.DeleteTask(&config.Config{})).Methods("DELETE")

			Convey("Task deleted successfully", func() {
				statusQueued := []eremetic.Status{
					eremetic.Status{
						Status: eremetic.TaskQueued,
						Time:   time.Now().Unix(),
					},
				}
				task := eremetic.Task{
					TaskCPUs:          0.2,
					TaskMem:           0.5,
					Command:           "test",
					Image:             "test",
					Status:            statusQueued,
					ID:                id,
					MaskedEnvironment: maskedEnv,
				}
				db.PutTask(&task)
				r.URL, _ = url.Parse(fmt.Sprintf("/task/%s", id))

				m.HandleFunc("/task/{task}", h.DeleteTask(&config.Config{}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusAccepted)

			})

			Convey("Task is in running state", func() {
				db.PutTask(&task)
				r.URL, _ = url.Parse(fmt.Sprintf("/task/%s", id))

				m.HandleFunc("/task/{task}", h.DeleteTask(&config.Config{}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusConflict)
			})

			Convey("Task not found", func() {
				id = "eremetic-task.4567"
				r.URL, _ = url.Parse(fmt.Sprintf("/task/%s", id))

				m.HandleFunc("/task/{task}", h.DeleteTask(&config.Config{}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusNotFound)
			})

		})

		Convey("Index", func() {
			r, _ := http.NewRequest("GET", "/", nil)

			Convey("Renders nothing for json", func() {
				m.HandleFunc("/", h.IndexHandler(&config.Config{}))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusNoContent)
			})

			Convey("Renders the html template for html requests when URLPrefix is empty", func() {
				conf := config.DefaultConfig()

				r.Header.Add("Accept", "text/html")
				m.HandleFunc("/", h.IndexHandler(conf))
				m.ServeHTTP(wr, r)

				b, _ := ioutil.ReadAll(wr.Body)
				body := string(b)

				So(wr.Code, ShouldEqual, http.StatusOK)
				So(body, ShouldContainSubstring, "<html>")
				So(body, ShouldContainSubstring, "<div id='eremetic-version'>test</div>")
				So(body, ShouldNotContainSubstring, "/service/eremetic")
			})

			Convey("Renders the html template for html requests when URLPrefix is set", func() {
				conf := config.DefaultConfig()
				conf.URLPrefix = "/service/eremetic"

				r.Header.Add("Accept", "text/html")
				m.HandleFunc("/", h.IndexHandler(conf))
				m.ServeHTTP(wr, r)

				b, _ := ioutil.ReadAll(wr.Body)
				body := string(b)

				So(wr.Code, ShouldEqual, http.StatusOK)
				So(body, ShouldContainSubstring, "<html>")
				So(body, ShouldContainSubstring, "<div id='eremetic-version'>test</div>")
				So(body, ShouldContainSubstring, "/service/eremetic")
			})
		})

	})
}
