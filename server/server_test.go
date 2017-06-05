package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/config"
	"github.com/eremetic-framework/eremetic/mock"
)

func TestServer(t *testing.T) {
	Convey("Server", t, func() {
		Convey("AddTask", func() {
			Convey("Simple", func() {
				sched := mock.Scheduler{
					ScheduleTaskFn: func(req eremetic.Request) (string, error) {
						return "task_id", nil
					},
				}

				db := mock.TaskDB{}
				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				var body bytes.Buffer
				body.WriteString(`{}`)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("POST", "http://example.com/task", &body)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusAccepted)
			})
			Convey("QueueFull", func() {
				sched := mock.Scheduler{
					ScheduleTaskFn: func(req eremetic.Request) (string, error) {
						return "", eremetic.ErrQueueFull
					},
				}

				db := mock.TaskDB{}
				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				var body bytes.Buffer
				body.WriteString(`{}`)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("POST", "http://example.com/task", &body)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusServiceUnavailable)
			})
			Convey("UnknownError", func() {
				sched := mock.Scheduler{
					ScheduleTaskFn: func(req eremetic.Request) (string, error) {
						return "", errors.New("unknown error")
					},
				}

				db := mock.TaskDB{}
				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				var body bytes.Buffer
				body.WriteString(`{}`)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("POST", "http://example.com/task", &body)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
		Convey("GetFromSandBox", func() {
			Convey("Simple", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("OK"))
				}))
				defer ts.Close()

				agentURL, _ := url.Parse(ts.URL)
				host := strings.Split(agentURL.Host, ":")
				port, _ := strconv.ParseInt(host[1], 10, 64)

				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ReadTaskFn: func(id string) (eremetic.Task, error) {
						return eremetic.Task{
							ID:          id,
							AgentIP:     host[0],
							AgentPort:   int32(port),
							SandboxPath: "/tmp",
						}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/task/test_id/stdout", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusOK)
				So(rec.Body.String(), ShouldEqual, "OK")
			})
			Convey("MissingSandboxPath", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ReadTaskFn: func(id string) (eremetic.Task, error) {
						return eremetic.Task{
							ID: id,
						}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/task/test_id/stdout", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusNoContent)
			})
		})
		Convey("GetTaskInfo", func() {
			Convey("Simple", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ReadTaskFn: func(id string) (eremetic.Task, error) {
						return eremetic.Task{
							ID: id,
						}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/task/test_id", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusOK)
			})
			Convey("TaskNotFound", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ReadTaskFn: func(id string) (eremetic.Task, error) {
						return eremetic.Task{}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/task/unknown_id", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusNotFound)
			})
		})
		Convey("ListRunningTasks", func() {
			Convey("Simple", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ListNonTerminalTasksFn: func() ([]*eremetic.Task, error) {
						return []*eremetic.Task{}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/task", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusOK)
			})
		})
		Convey("Index", func() {
			Convey("Simple", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ListNonTerminalTasksFn: func() ([]*eremetic.Task, error) {
						return []*eremetic.Task{}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/", nil)
				r.Header.Set("Accept", "text/html")

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusOK)
			})
			Convey("DoesNotAcceptHTML", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ListNonTerminalTasksFn: func() ([]*eremetic.Task, error) {
						return []*eremetic.Task{}, nil
					},
				}

				cfg := config.Config{}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/", nil)

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusNoContent)
			})
		})
		Convey("Auth", func() {
			Convey("IndexUnauthorized", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ListNonTerminalTasksFn: func() ([]*eremetic.Task, error) {
						return []*eremetic.Task{}, nil
					},
				}

				cfg := config.Config{HTTPCredentials: "admin:admin"}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/", nil)
				r.Header.Set("Accept", "text/html")

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("IndexOK", func() {
				sched := mock.Scheduler{}

				db := mock.TaskDB{
					ListNonTerminalTasksFn: func() ([]*eremetic.Task, error) {
						return []*eremetic.Task{}, nil
					},
				}

				cfg := config.Config{HTTPCredentials: "admin:admin"}

				srv := NewRouter(&sched, &cfg, &db)

				rec := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://example.com/", nil)
				r.Header.Set("Accept", "text/html")
				r.SetBasicAuth("admin", "admin")

				srv.ServeHTTP(rec, r)

				So(rec.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
