package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/config"
)

// Route enforces the structure of a route
type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

// Routes is a collection of route structs
type Routes []Route

// NewRouter is used to create a new router.
func NewRouter(scheduler eremetic.Scheduler, conf *config.Config, db eremetic.TaskDB) *mux.Router {
	h := NewHandler(scheduler, db)
	router := mux.NewRouter().StrictSlash(true)

	if conf.URLPrefix != "" {
		logrus.Debugf("Injecting URLPrefix path prefix: %s", conf.URLPrefix)
		subrouter := router.PathPrefix(conf.URLPrefix).Subrouter()
		router = subrouter
	}

	for _, route := range routes(h, conf) {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(prometheus.InstrumentHandler(route.Name, route.Handler))
	}

	router.
		PathPrefix("/static/").
		Handler(h.StaticAssets(conf))

	router.NotFoundHandler = http.HandlerFunc(h.NotFound(conf))

	username, password := parseHTTPCredentials(conf.HTTPCredentials)
	if username != "" && password != "" {
		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			name := route.GetName()
			// `/version` can be used as health check, so ignore auth required for it
			if name != "Version" {
				route.Handler(authWrap(route.GetHandler(), username, password))
			}
			return nil
		})
	}

	// Print out routes for debug purposes
	logrus.Debugf("Registered route handlers:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			logrus.Debugf("Unable to list routes.")
		}
		logrus.Debugf(t)
		return nil
	})

	return router
}

func routes(h Handler, conf *config.Config) Routes {
	return Routes{
		Route{
			Name:    "AddTask",
			Method:  "POST",
			Pattern: "/task",
			Handler: h.AddTask(conf),
		},
		Route{
			Name:    "Status",
			Method:  "GET",
			Pattern: "/task/{taskId}",
			Handler: h.GetTaskInfo(conf),
		},
		Route{
			Name:    "STDOUT",
			Method:  "GET",
			Pattern: "/task/{taskId}/stdout",
			Handler: h.GetFromSandbox("stdout"),
		},
		Route{
			Name:    "STDERR",
			Method:  "GET",
			Pattern: "/task/{taskId}/stderr",
			Handler: h.GetFromSandbox("stderr"),
		},
		Route{
			Name:    "Kill",
			Method:  "POST",
			Pattern: "/task/{taskId}/kill",
			Handler: h.KillTask(conf),
		},
		Route{
			Name:    "Delete",
			Method:  "DELETE",
			Pattern: "/task/{taskId}",
			Handler: h.DeleteTask(conf),
		},
		Route{
			Name:    "ListRunningTasks",
			Method:  "GET",
			Pattern: "/task",
			Handler: h.ListRunningTasks(),
		},
		Route{
			Name:    "Index",
			Method:  "GET",
			Pattern: "/",
			Handler: h.IndexHandler(conf),
		},
		Route{
			Name:    "Version",
			Method:  "GET",
			Pattern: "/version",
			Handler: h.Version(conf),
		},
		Route{
			Name:    "Metrics",
			Method:  "GET",
			Pattern: "/metrics",
			Handler: prometheus.Handler(),
		},
	}
}
