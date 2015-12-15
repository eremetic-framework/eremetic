package routes

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/klarna/eremetic/assets"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
	log "github.com/dmuth/google-go-log4go"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
)

// Create is used to create a new router
func Create(scheduler types.Scheduler) *mux.Router {
	routes := types.Routes{
		types.Route{
			Name:        "AddTask",
			Method:      "POST",
			Pattern:     "/task",
			HandlerFunc: handler.AddTask(scheduler),
		},
		types.Route{
			Name:        "Status",
			Method:      "GET",
			Pattern:     "/task/{taskId}",
			HandlerFunc: handler.GetTaskInfo(scheduler),
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	router.PathPrefix("/static/").
		Handler(
		http.StripPrefix(
			"/static/", http.FileServer(
				&assetfs.AssetFS{Asset: assets.Asset, AssetDir: assets.AssetDir, Prefix: "static"})))

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
