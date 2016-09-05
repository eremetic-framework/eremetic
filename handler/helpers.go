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
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/formatter"
	"github.com/klarna/eremetic/types"
)

// getFile handles the actual fetching of file from the agent.
func getFile(file string, task types.EremeticTask) (int, io.ReadCloser) {
	if task.SandboxPath == "" {
		return http.StatusNoContent, nil
	}

	url := fmt.Sprintf(
		"http://%s:%d/files/download?path=%s/%s",
		task.AgentIP,
		task.AgentPort,
		task.SandboxPath,
		file,
	)

	logrus.WithField("url", url).Debug("Fetching file from sandbox")

	response, err := http.Get(url)

	if err != nil {
		logrus.WithError(err).Errorf("Unable to fetch %s from agent %s.", file, task.SlaveId)
		return http.StatusInternalServerError, ioutil.NopCloser(strings.NewReader("Unable to fetch upstream file."))
	}

	return http.StatusOK, response.Body
}

func handleError(err error, w http.ResponseWriter, message string) {
	if err == nil {
		return
	}

	errorMessage := ErrorDocument{
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

func renderHTML(w http.ResponseWriter, r *http.Request, task types.EremeticTask, taskID string, conf *config.Config) {
	var templateFile string

	data := make(map[string]interface{})
	funcMap := template.FuncMap{
		"ToLower":    strings.ToLower,
		"FormatTime": formatter.FormatTime,
	}

	if reflect.DeepEqual(task, (types.EremeticTask{})) {
		notFound(w, r)
		return
	} else {
		templateFile = "task.html"
		data = makeMap(task)
		data["Version"] = conf.Version
	}

	source, _ := assets.Asset(fmt.Sprintf("templates/%s", templateFile))
	tpl, err := template.New(templateFile).Funcs(funcMap).Parse(string(source))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.WithError(err).WithField("template", templateFile).Error("Unable to load template")
		return
	}

	err = tpl.Execute(w, data)
	if err != nil {
		logrus.WithError(err).WithField("template", templateFile).Error("Unable to execute template")
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		src, _ := assets.Asset("templates/error_404.html")
		tpl, err := template.New("404").Parse(string(src))
		if err != nil {
			logrus.WithError(err).WithField("template", "error_404.html").Error("Unable to load template")
		}
		err = tpl.Execute(w, nil)
		if err != nil {
			logrus.WithError(err).WithField("template", "error_404.html").Error("Unable to execute template")
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(nil)
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
	data["SlaveConstraints"] = task.SlaveConstraints
	data["Status"] = task.Status
	data["CPU"] = fmt.Sprintf("%.2f", task.TaskCPUs)
	data["Memory"] = fmt.Sprintf("%.2f", task.TaskMem)

	return data
}

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
