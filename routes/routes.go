package routes

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Create is used to create a new router
func Create(scheduler types.Scheduler, conf *config.Config) *mux.Router {
	h := handler.Create(scheduler, conf.Database)
	routes := types.Routes{
		types.Route{
			Name:    "AddTask",
			Method:  "POST",
			Pattern: "/task",
			Handler: h.AddTask(),
		},
		types.Route{
			Name:    "Status",
			Method:  "GET",
			Pattern: "/task/{taskId}",
			Handler: h.GetTaskInfo(conf),
		},
		types.Route{
			Name:    "STDOUT",
			Method:  "GET",
			Pattern: "/task/{taskId}/stdout",
			Handler: h.GetFromSandbox("stdout"),
		},
		types.Route{
			Name:    "STDERR",
			Method:  "GET",
			Pattern: "/task/{taskId}/stderr",
			Handler: h.GetFromSandbox("stderr"),
		},
		types.Route{
			Name:    "ListRunningTasks",
			Method:  "GET",
			Pattern: "/task",
			Handler: h.ListRunningTasks(),
		},
		types.Route{
			Name:    "ListFinishedTasks",
			Method:  "GET",
			Pattern: "/history",
			Handler: h.ListTerminatedTasks(),
		},
		types.Route{
			Name:    "Index",
			Method:  "GET",
			Pattern: "/",
			Handler: h.IndexHandler(conf),
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
		Path("/version").
		Name("Version").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(conf.Version)
		})

	router.PathPrefix("/static/").
		Handler(
			http.StripPrefix(
				"/static/", http.FileServer(
					&assetfs.AssetFS{Asset: assets.Asset, AssetDir: assets.AssetDir, AssetInfo: assets.AssetInfo, Prefix: "static"})))

	router.NotFoundHandler = http.HandlerFunc(notFound)

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
