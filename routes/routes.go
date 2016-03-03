package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

// Create is used to create a new router
func Create(scheduler types.Scheduler) *mux.Router {
	routes := types.Routes{
		types.Route{
			Name:    "AddTask",
			Method:  "POST",
			Pattern: "/task",
			Handler: handler.AddTask(scheduler),
		},
		types.Route{
			Name:    "Status",
			Method:  "GET",
			Pattern: "/task/{taskId}",
			Handler: handler.GetTaskInfo(scheduler),
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(prometheus.InstrumentHandler(route.Name, route.Handler))
	}

	router.
		Methods("GET").
		Path("/metrics").
		Name("Metrics").
		Handler(prometheus.Handler())

	router.
		Methods("GET").
		Path("/").
		Name("Index").
		HandlerFunc(indexHandler)

	router.PathPrefix("/static/").
		Handler(
			http.StripPrefix(
				"/static/", http.FileServer(
					&assetfs.AssetFS{Asset: assets.Asset, AssetDir: assets.AssetDir, AssetInfo: assets.AssetInfo, Prefix: "static"})))

	router.NotFoundHandler = authWrap(http.HandlerFunc(notFound))

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		route.Handler(authWrap(route.GetHandler()))
		return nil
	})

	return router
}

func notFound(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		src, _ := assets.Asset("templates/error_404.html")
		tpl, err := template.New("404").Parse(string(src))
		if err == nil {
			tpl.Execute(w, nil)
			return
		}
		logrus.WithError(err).WithField("template", "error_404.html").Error("Unable to load template")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		src, _ := assets.Asset("templates/index.html")
		tpl, err := template.New("index").Parse(string(src))
		if err == nil {
			tpl.Execute(w, nil)
			return
		}
		logrus.WithError(err).WithField("template", "index.html").Error("Unable to load template")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(nil)
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

func authWrap(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http_credentials := viper.GetString("http_credentials")
		if http_credentials != "" {
			pair := strings.SplitN(http_credentials, ":", 2)
			if len(pair) == 2 {
				err := checkAuth(r, pair[0], pair[1])
				if err != nil {
					requireAuth(w, r)
					return
				}
			} else {
				logrus.WithField("http_credentials", http_credentials).Error("using 'username:password' format for http_credentials")
			}
		}
		fn.ServeHTTP(w, r)
	}
}
