package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alde/eremetic/database"
	"github.com/alde/eremetic/types"
	"github.com/gorilla/mux"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

type mockError struct {
	message string
}

func (m mockError) Error() string {
	return m.message
}

type mockScheduler struct {
}

func (s *mockScheduler) ScheduleTask(request types.Request) (string, error) {
	return "eremetic-task.mock", nil
}

func TestHandling(t *testing.T) {
	scheduler := &mockScheduler{}
	status := []types.Status{
		types.Status{
			Status: mesos.TaskState_TASK_RUNNING.String(),
			Time:   time.Now().Unix(),
		},
	}

	dir, _ := os.Getwd()
	database.NewDB(fmt.Sprintf("%s/../db/test.db", dir))
	database.Clean()
	defer database.Close()

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

			handleError(err, wr)

			So(wr.Code, ShouldEqual, 422)
			So(strings.TrimSpace(wr.Body.String()), ShouldEqual, "{}")
		})
	})

	Convey("GetTaskInfo", t, func() {
		wr := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)
		m := mux.NewRouter()
		m.HandleFunc("/task/{taskId}", GetTaskInfo(scheduler))

		Convey("Not Found", func() {
			id := "eremetic-task.5678"
			task := types.EremeticTask{
				TaskCPUs:  0.2,
				TaskMem:   0.5,
				Command:   &mesos.CommandInfo{},
				Container: &mesos.ContainerInfo{},
				Status:    status,
				ID:        id,
			}
			database.PutTask(&task)
			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Found", func() {
			id := "eremetic-task.1234"
			task := types.EremeticTask{
				TaskCPUs:  0.2,
				TaskMem:   0.5,
				Command:   &mesos.CommandInfo{},
				Container: &mesos.ContainerInfo{},
				Status:    status,
				ID:        id,
			}
			database.PutTask(&task)
			m.ServeHTTP(wr, r)

			So(wr.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("AddTask", t, func() {
		wr := httptest.NewRecorder()

		Convey("It should respond with a location header", func() {
			data := []byte(`{"task_mem":22.0, "docker_image": "busybox", "command": "echo hello", "task_cpus":0.5, "tasks_to_launch": 1}`)
			r, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(data))

			handler := AddTask(scheduler)
			handler(wr, r)

			location := wr.HeaderMap["Location"][0]
			So(location, ShouldStartWith, "/task/eremetic-task.")
			So(wr.Code, ShouldEqual, http.StatusAccepted)
		})
	})
}
