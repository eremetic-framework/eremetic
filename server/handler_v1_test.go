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
	"github.com/eremetic-framework/eremetic/api"
	"github.com/eremetic-framework/eremetic/config"
	"github.com/eremetic-framework/eremetic/mock"
	"github.com/eremetic-framework/eremetic/version"
)

func TestHandlingV1(t *testing.T) {
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
		r, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/task/%s", id), nil)

		Convey("GetTaskInfo", func() {
			m.HandleFunc("/api/v1/task/{taskId}", h.GetTaskInfo(&config.Config{}, api.V1))

			Convey("JSON request", func() {
				Convey("Not Found", func() {
					id = "eremetic-task.5678"
					r.URL, _ = url.Parse(fmt.Sprintf("/api/v1/task/%s", id))

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

					var retrievedTask api.TaskV1
					body, _ := ioutil.ReadAll(io.LimitReader(wr.Body, 1048576))
					json.Unmarshal(body, &retrievedTask)

					So(retrievedTask.MaskedEnvironment, ShouldContainKey, "foo")
					So(retrievedTask.MaskedEnvironment["foo"], ShouldNotEqual, "bar")
					So(retrievedTask.MaskedEnvironment["foo"], ShouldEqual, "*******")

				})
			})
		})

		Convey("AddTask", func() {
			data := []byte(`{"mem":22.0, "image": "busybox", "command": "echo hello", "cpu":0.5}`)
			r, _ := http.NewRequest("POST", "/api/v1/task", bytes.NewBuffer(data))
			r.Host = "localhost"

			Convey("It should respond with a location header", func() {
				handler := h.AddTask(&config.Config{}, api.V1)
				handler(wr, r)

				location := wr.HeaderMap["Location"][0]
				So(location, ShouldStartWith, "http://localhost/api/v1/task/eremetic-task.")
				So(wr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("It should respond with a location per URL prefix header", func() {
				conf := config.Config{}
				conf.URLPrefix = "/service/eremetic"
				handler := h.AddTask(&conf, api.V1)
				handler(wr, r)

				location := wr.HeaderMap["Location"][0]
				So(location, ShouldStartWith, "http://localhost/service/eremetic/api/v1/task/eremetic-task.")
				So(wr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("Failed to schedule", func() {
				err := errors.New("A random error")
				scheduler.NextError = &err

				handler := h.AddTask(&config.Config{}, api.V1)
				handler(wr, r)

				So(wr.Code, ShouldEqual, 500)
			})

			Convey("Error on bad input stream", func() {
				r.Body = ioutil.NopCloser(&mock.ErrorReader{})

				handler := h.AddTask(&config.Config{}, api.V1)
				handler(wr, r)

				So(wr.Code, ShouldEqual, 422)
			})

			Convey("Error on malformed json", func() {
				data = []byte(`{"key:123}`)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

				handler := h.AddTask(&config.Config{}, api.V1)
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
					r, _ := http.NewRequest("GET", "/api/v1/task/eremetic-task.1234/stdout", nil)
					m.HandleFunc("/api/v1/task/{taskId}/stdout", h.GetFromSandbox("stdout", api.V1))
					m.ServeHTTP(wr, r)

					So(wr.Code, ShouldEqual, http.StatusOK)
					So(wr.Header().Get("Content-Type"), ShouldEqual, "text/plain; charset=UTF-8")

					body, _ := ioutil.ReadAll(wr.Body)
					So(string(body), ShouldEqual, "mocked")
				})

				Convey("stderr", func() {
					r, _ := http.NewRequest("GET", "/api/v1/task/eremetic-task.1234/stderr", nil)
					m.HandleFunc("/api/v1/task/{taskId}/stderr", h.GetFromSandbox("stderr", api.V1))

					m.ServeHTTP(wr, r)

					So(wr.Code, ShouldEqual, http.StatusOK)
					So(wr.Header().Get("Content-Type"), ShouldEqual, "text/plain; charset=UTF-8")

					body, _ := ioutil.ReadAll(wr.Body)
					So(string(body), ShouldEqual, "mocked")
				})

				Convey("No Sandbox path", func() {
					r, _ := http.NewRequest("GET", "/api/v1/task/eremetic-task.1234/stdout", nil)
					m.HandleFunc("/api/v1/task/{taskId}/stdout", h.GetFromSandbox("stdout", api.V1))

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
			m.HandleFunc("/version", h.Version(&config.Config{}, api.V1))
			m.ServeHTTP(wr, r)

			body, _ := ioutil.ReadAll(wr.Body)
			So(string(body), ShouldEqual, "\"test\"\n")
		})

		Convey("Delete task", func() {
			r, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/task/%s", id), nil)
			m.HandleFunc("/api/v1/task/{taskId}", h.DeleteTask(&config.Config{}, api.V1)).Methods("DELETE")

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
				r.URL, _ = url.Parse(fmt.Sprintf("/api/v1/task/%s", id))

				m.HandleFunc("/api/v1/task/{task}", h.DeleteTask(&config.Config{}, api.V1))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusAccepted)

			})

			Convey("Task is in running state", func() {
				db.PutTask(&task)
				r.URL, _ = url.Parse(fmt.Sprintf("/api/v1/task/%s", id))

				m.HandleFunc("/api/v1/task/{task}", h.DeleteTask(&config.Config{}, api.V1))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusConflict)
			})

			Convey("Task not found", func() {
				id = "eremetic-task.4567"
				r.URL, _ = url.Parse(fmt.Sprintf("/api/v1/task/%s", id))

				m.HandleFunc("/api/v1/task/{task}", h.DeleteTask(&config.Config{}, api.V1))
				m.ServeHTTP(wr, r)

				So(wr.Code, ShouldEqual, http.StatusNotFound)
			})

		})
	})
}
