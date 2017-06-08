package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/config"
	"github.com/eremetic-framework/eremetic/server/assets"
	"github.com/eremetic-framework/eremetic/version"
)

// getFile handles the actual fetching of file from the agent.
func getFile(file string, task eremetic.Task) (int, io.ReadCloser) {
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
		logrus.WithError(err).Errorf("Unable to fetch %s from agent %s.", file, task.AgentID)
		return http.StatusInternalServerError, ioutil.NopCloser(strings.NewReader("Unable to fetch upstream file."))
	}

	return http.StatusOK, response.Body
}

func handleError(err error, w http.ResponseWriter, message string) {
	if err == nil {
		return
	}

	errorMessage := errorDocument{
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

func renderHTML(w http.ResponseWriter, r *http.Request, task eremetic.Task, taskID string, conf *config.Config) {
	var templateFile string

	data := make(map[string]interface{})
	funcMap := template.FuncMap{
		"ToLower":    strings.ToLower,
		"FormatTime": FormatTime,
	}

	if reflect.DeepEqual(task, (eremetic.Task{})) {
		notFound(w, r, conf)
		return
	}

	templateFile = "task.html"
	data = makeMap(task)
	data["Version"] = version.Version
	data["URLPrefix"] = conf.URLPrefix

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

func notFound(w http.ResponseWriter, r *http.Request, conf *config.Config) {
	w.WriteHeader(http.StatusNotFound)

	data := make(map[string]interface{})
	data["URLPrefix"] = conf.URLPrefix

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		src, _ := assets.Asset("templates/error_404.html")
		tpl, err := template.New("404").Parse(string(src))
		if err != nil {
			logrus.WithError(err).WithField("template", "error_404.html").Error("Unable to load template")
		}
		err = tpl.Execute(w, data)
		if err != nil {
			logrus.WithError(err).WithField("template", "error_404.html").Error("Unable to execute template")
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(nil)
}

func makeMap(task eremetic.Task) map[string]interface{} {
	data := make(map[string]interface{})

	data["TaskID"] = task.ID
	data["CommandEnv"] = task.Environment
	data["CommandUser"] = task.User
	data["Command"] = task.Command
	// TODO: Support more than docker?
	data["ContainerImage"] = task.Image
	data["FrameworkID"] = task.FrameworkID
	data["Hostname"] = task.Hostname
	data["Name"] = task.Name
	data["AgentID"] = task.AgentID
	data["AgentConstraints"] = task.AgentConstraints
	data["Status"] = task.Status
	data["CPU"] = fmt.Sprintf("%.2f", task.TaskCPUs)
	data["Memory"] = fmt.Sprintf("%.2f", task.TaskMem)
	data["Terminated"] = task.IsTerminated()

	return data
}

func absURL(r *http.Request, path string, conf *config.Config) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	if conf.URLPrefix != "" {
		path = fmt.Sprintf("%s%s", conf.URLPrefix, path)
		logrus.WithField("path", path).Debug("absurl was computed")
	}

	url := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   path,
	}
	return url.String()
}

func parseHTTPCredentials(credentials string) (string, string) {
	if credentials == "" {
		return "", ""
	}

	pair := strings.SplitN(credentials, ":", 2)
	if len(pair) != 2 {
		logrus.WithField("http_credentials", credentials).Error("using 'username:password' format for http_credentials")
		return "", ""
	}

	return pair[0], pair[1]
}

func checkAuth(r *http.Request, user string, password string) error {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	badErr := errors.New("bad authorization")

	if len(s) != 2 || s[0] != "Basic" {
		return badErr
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return err
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return badErr
	}
	if pair[0] != user || pair[1] != password {
		return badErr
	}
	return nil
}

func requireAuth(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		src, _ := assets.Asset("templates/error_401.html")
		tpl, err := template.New("401").Parse(string(src))
		if err == nil {
			w.Header().Set("WWW-Authenticate", `basic realm="Eremetic"`)
			w.WriteHeader(http.StatusUnauthorized)
			tpl.Execute(w, nil)
			return
		}
		logrus.WithError(err).WithField("template", "error_401.html").Error("Unable to load template")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(nil)
}

func authWrap(fn http.Handler, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := checkAuth(r, username, password)
		if err != nil {
			requireAuth(w, r)
			return
		}

		fn.ServeHTTP(w, r)
	}
}
