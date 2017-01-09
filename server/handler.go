package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"

	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/server/assets"
	"github.com/klarna/eremetic/version"
)

type errorDocument struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Handler holds the server context.
type Handler struct {
	scheduler eremetic.Scheduler
	database  eremetic.TaskDB
}

// NewHandler returns a new instance of Handler.
func NewHandler(scheduler eremetic.Scheduler, database eremetic.TaskDB) Handler {
	return Handler{
		scheduler: scheduler,
		database:  database,
	}
}

// AddTask handles adding a task to the queue
func (h Handler) AddTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request eremetic.Request

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
			if err == eremetic.ErrQueueFull {
				httpStatus = 503
			}
			errorMessage := errorDocument{
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
func (h Handler) GetTaskInfo(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		logrus.WithField("task_id", id).Debug("Fetching task")
		task, _ := h.database.ReadTask(id)

		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			renderHTML(w, r, task, id, conf)
		} else {
			if reflect.DeepEqual(task, (eremetic.Task{})) {
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
			return
		}
		writeJSON(200, tasks, w)
	}
}

// IndexHandler returns the index template, or no content.
func (h Handler) IndexHandler(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			src, _ := assets.Asset("templates/index.html")
			tpl, err := template.New("index").Parse(string(src))
			data := make(map[string]interface{})
			data["Version"] = version.Version
			if err == nil {
				tpl.Execute(w, data)
				return
			}
			logrus.WithError(err).WithField("template", "index.html").Error("Unable to load template")
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(nil)
	}
}

// Version returns the currently running Eremetic version.
func (h Handler) Version(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(version.Version)
	}
}

// NotFound is in charge of reporting that a task can not be found.
func (h Handler) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Proxy to the notFound helper function
		notFound(w, r)
	}
}

// StaticAssets handles the serving of compiled static assets.
func (h Handler) StaticAssets() http.Handler {
	return http.StripPrefix(
		"/static/", http.FileServer(
			&assetfs.AssetFS{Asset: assets.Asset, AssetDir: assets.AssetDir, AssetInfo: assets.AssetInfo, Prefix: "static"}))
}

func (h Handler) KillTask(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		logrus.WithField("task_id", id).Debug("Killing task")
		err := h.scheduler.Kill(id)
		respStatus := http.StatusAccepted
		var body string
		if err != nil {
			respStatus = http.StatusInternalServerError
			body = err.Error()
		}
		writeJSON(respStatus, body, w)
	}
}

func (h Handler) DeleteTask(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		logrus.WithField("task_id", id).Debug("Deleting task")
		respStatus := http.StatusAccepted
		var body string
		task, err := h.database.ReadTask(id)
		if (err != nil) {
			respStatus = http.StatusNotFound
			writeJSON(respStatus, err.Error(), w)
			return
		}
		if (task.IsRunning()) {
			respStatus = http.StatusConflict
			errMsg := fmt.Sprintf("Cannot delete the task [%s]. As it is still running.", id)
			logrus.WithField("task_id", id).Debug(errMsg)
			writeJSON(respStatus, errMsg, w)
			return
		}
		err = h.database.DeleteTask(id)
		if err != nil {
			respStatus = http.StatusInternalServerError
			body = err.Error()
		}
		writeJSON(respStatus, body, w)
	}
}
