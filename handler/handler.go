package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/dmuth/google-go-log4go"
	"github.com/gorilla/mux"

	"github.com/alde/eremetic/types"
)

var requests = make(chan *types.Request)
var scheduler *eremeticScheduler

// AddTask handles adding a task to the queue
func AddTask(w http.ResponseWriter, r *http.Request) {
	var request types.Request

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	handleError(err, w)

	err = json.Unmarshal(body, &request)
	handleError(err, w)

	taskID, err := scheduleTask(scheduler, request)
	if err != nil {
		writeJSON(500, err, w)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/task/%s", taskID))
	writeJSON(http.StatusAccepted, taskID, w)
}

// GetTaskInfo returns information about the given task.
func GetTaskInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["taskId"]
	log.Debugf("Fetching task for id: %s", id)
	task := runningTasks[id]

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		renderHTML(w, r, task, id)
	} else {
		if task == (eremeticTask{}) {
			writeJSON(http.StatusNotFound, nil, w)
			return
		}
		writeJSON(http.StatusOK, task, w)
	}
}

// Run the RequestChannel Listener
func Run() {
	runningTasks = make(map[string]eremeticTask)
	scheduler = createEremeticScheduler()
	driver, err := createDriver(scheduler)

	if err != nil {
		log.Errorf("Unable to create scheduler driver: %s", err)
		return
	}

	defer close(scheduler.shutdown)
	defer driver.Stop(false)

	go func() {
		if status, err := driver.Run(); err != nil {
			log.Errorf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
		}
		log.Info("Exiting...")
	}()

	log.Debug("Entering handler.Run loop")
	for {
		select {
		case req := <-requests:
			log.Debug("Found a request in the queue!")
			scheduleTask(scheduler, *req)
		}
	}
}

// CleanupTasks is an infinite loop removing terminal Tasks that have stuck around too long.
func CleanupTasks() {
	for {
		for i, t := range runningTasks {
			if t.deleteAt.After(time.Now()) && types.IsTerminalString(t.Status) {
				delete(runningTasks, i)
			}
		}
		time.Sleep(time.Minute * 15)
	}
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

func renderHTML(w http.ResponseWriter, r *http.Request, task eremeticTask, taskID string) {
	var err error
	var tpl *template.Template

	data := make(map[string]interface{})
	funcMap := template.FuncMap{
		"Label": LabelColor,
	}

	if task == (eremeticTask{}) {
		tpl, err = template.ParseFiles("templates/error_404.html")
		data["TaskID"] = taskID
	} else {
		tpl, err = template.New("task.html").Funcs(funcMap).ParseFiles("templates/task.html")
		data = makeMap(task)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	err = tpl.Execute(w, data)
}

func makeMap(task eremeticTask) map[string]interface{} {
	data := make(map[string]interface{})

	data["TaskID"] = task.ID
	data["CommandEnv"] = task.Command.GetEnvironment().GetVariables()
	data["CommandUser"] = task.Command.GetUser()
	data["Command"] = task.Command.GetValue()
	// TODO: Support more than docker?
	data["ContainerImage"] = task.Container.GetDocker().GetImage()
	data["FrameworkID"] = task.FrameworkId
	data["Hostname"] = task.Hostname
	data["Name"] = task.Name
	data["SlaveID"] = task.SlaveId
	data["Status"] = task.Status
	data["CPU"] = fmt.Sprintf("%.2f", task.TaskCPUs)
	data["Memory"] = fmt.Sprintf("%.2f", task.TaskMem)

	return data
}

// LabelColor is used to map a status to a label color
func LabelColor(status string) string {
	switch status {
	case "TASK_FAILED":
		return "red"
	case "TASK_LOST":
		return "purple"
	case "TASK_KILLED":
		return "orange"
	case "TASK_FINISHED":
		return "green"
	default:
		return ""
	}
}
