package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	log "github.com/dmuth/google-go-log4go"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/formatter"
	"github.com/klarna/eremetic/types"
)

// AddTask handles adding a task to the queue
func AddTask(scheduler types.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request types.Request

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		handleError(err, w)

		err = json.Unmarshal(body, &request)
		handleError(err, w)

		taskID, err := scheduler.ScheduleTask(request)
		if err != nil {
			writeJSON(500, err, w)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/task/%s", taskID))
		writeJSON(http.StatusAccepted, taskID, w)
	}
}

// GetTaskInfo returns information about the given task.
func GetTaskInfo(scheduler types.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["taskId"]
		log.Debugf("Fetching task for id: %s", id)
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

type callbackData struct {
	Time   int64  `json:"time"`
	Status string `json:"status"`
	TaskID string `json:"task_id"`
}

// NotifyCallback handles posting a JSON back to the URI given with the task.
func NotifyCallback(task *types.EremeticTask) {
	if len(task.CallbackURI) == 0 {
		return
	}

	cbData := callbackData{
		Time:   task.Status[len(task.Status)-1].Time,
		Status: task.Status[len(task.Status)-1].Status,
		TaskID: task.ID,
	}

	body, err := json.Marshal(cbData)
	if err != nil {
		log.Errorf("Unable to create message for task %s, target uri %s", task.ID, task.CallbackURI)
		return
	}

	go func() {
		_, err = http.Post(task.CallbackURI, "application/json", bytes.NewBuffer(body))

		if err != nil {
			log.Error(err.Error())
		} else {
			log.Debugf("Sent callback to %s", task.CallbackURI)
		}
	}()

}

func handleError(err error, w http.ResponseWriter) {
	if err != nil {
		if err = writeJSON(422, err, w); err != nil {
			panic(err)
		}
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
		log.Error(err.Error())
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
