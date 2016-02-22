package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/formatter"
	"github.com/klarna/eremetic/types"
)

func absURL(r *http.Request, path string) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	url := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   path,
	}
	return url.String()
}

// AddTask handles adding a task to the queue
func AddTask(scheduler types.Scheduler) http.HandlerFunc {
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

		taskID, err := scheduler.ScheduleTask(request)
		if err != nil {
			writeJSON(500, err, w)
			return
		}

		w.Header().Set("Location", absURL(r, fmt.Sprintf("/task/%s", taskID)))
		writeJSON(http.StatusAccepted, taskID, w)
	}
}

// GetTaskInfo returns information about the given task.
func GetTaskInfo(scheduler types.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		logrus.WithField("task_id", id).Debug("Fetching task")
		task, _ := database.ReadTask(id)

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

func handleError(err error, w http.ResponseWriter, message string) {
	if err == nil {
		return
	}

	var errorMessage = struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		err.Error(),
		message,
	}

	if err = writeJSON(422, errorMessage, w); err != nil {
		logrus.WithError(err).WithField("message", message).Panic("Unable to respond")
	}
}

func writeJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func renderHTML(w http.ResponseWriter, r *http.Request, task types.EremeticTask, taskID string) {
	var templateFile string

	data := make(map[string]interface{})
	funcMap := template.FuncMap{
		"ToLower":    strings.ToLower,
		"FormatTime": formatter.FormatTime,
	}

	if reflect.DeepEqual(task, (types.EremeticTask{})) {
		templateFile = "error_404.html"
		data["TaskID"] = taskID
	} else {
		templateFile = "task.html"
		data = makeMap(task)
	}

	source, _ := assets.Asset(fmt.Sprintf("templates/%s", templateFile))
	tpl, err := template.New(templateFile).Funcs(funcMap).Parse(string(source))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.WithError(err).WithField("template", templateFile).Error("Unable to render template")
		return
	}

	err = tpl.Execute(w, data)
}

func makeMap(task types.EremeticTask) map[string]interface{} {
	data := make(map[string]interface{})

	data["TaskID"] = task.ID
	data["CommandEnv"] = task.Environment
	data["CommandUser"] = task.User
	data["Command"] = task.Command
	// TODO: Support more than docker?
	data["ContainerImage"] = task.Image
	data["FrameworkID"] = task.FrameworkId
	data["Hostname"] = task.Hostname
	data["Name"] = task.Name
	data["SlaveID"] = task.SlaveId
	data["Status"] = task.Status
	data["CPU"] = fmt.Sprintf("%.2f", task.TaskCPUs)
	data["Memory"] = fmt.Sprintf("%.2f", task.TaskMem)

	return data
}
