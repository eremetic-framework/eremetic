package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/scheduler"
	"github.com/klarna/eremetic/types"
)

type ErrorDocument struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Handler struct {
	scheduler types.Scheduler
	database  database.TaskDB
}

func Create(scheduler types.Scheduler, database database.TaskDB) Handler {
	return Handler{
		scheduler: scheduler,
		database:  database,
	}
}

// AddTask handles adding a task to the queue
func (h Handler) AddTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request types.Request

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			handleError(err, w, "Unable to read payload.")
			return
		}

		err = json.Unmarshal(body, &request)
		if err != nil {
			handleError(err, w, "Unable to parse body into a valid request.")
			return
		}

		taskID, err := h.scheduler.ScheduleTask(request)
		if err != nil {
			logrus.WithError(err).Error("Unable to create task.")
			httpStatus := 500
			if err == scheduler.ErrQueueFull {
				httpStatus = 503
			}
			errorMessage := ErrorDocument{
				err.Error(),
				"Unable to schedule task",
			}
			writeJSON(httpStatus, errorMessage, w)
			return
		}

		w.Header().Set("Location", absURL(r, fmt.Sprintf("/task/%s", taskID)))
		writeJSON(http.StatusAccepted, taskID, w)
	}
}

// GetFromSandbox fetches a file from the sandbox of the agent that ran the task
func (h Handler) GetFromSandbox(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		taskID := vars["taskId"]
		task, _ := h.database.ReadTask(taskID)

		status, data := getFile(file, task)

		if status != http.StatusOK {
			writeJSON(status, data, w)
			return
		}

		defer data.Close()
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, data)
	}
}

// GetTaskInfo returns information about the given task.
func (h Handler) GetTaskInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		logrus.WithField("task_id", id).Debug("Fetching task")
		task, _ := h.database.ReadTask(id)

		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			renderHTML(w, r, task, id)
		} else {
			if reflect.DeepEqual(task, (types.EremeticTask{})) {
				writeJSON(http.StatusNotFound, nil, w)
				return
			}
			writeJSON(http.StatusOK, task, w)
		}
	}
}

// ListRunningTasks returns information about running tasks in the database.
func (h Handler) ListRunningTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("Fetching all tasks")
		tasks, err := h.database.ListNonTerminalTasks()
		if err != nil {
			handleError(err, w, "Unable to fetch running tasks from the database")
		}
		writeJSON(200, tasks, w)
	}
}
