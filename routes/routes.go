package routes

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	log "github.com/dmuth/google-go-log4go"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
	"github.com/prometheus/client_golang/prometheus"
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
		log.Error(err.Error())
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(nil)
}
