package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Create is used to create a new router
func Create(scheduler types.Scheduler, conf *config.Config, db database.TaskDB) *mux.Router {
	h := handler.Create(scheduler, db)
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes(h, conf) {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(prometheus.InstrumentHandler(route.Name, route.Handler))
	}

	router.
		PathPrefix("/static/").
		Handler(h.StaticAssets())

	router.NotFoundHandler = http.HandlerFunc(h.NotFound())

	return router
}

func routes(h handler.Handler, conf *config.Config) types.Routes {
	return types.Routes{
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
			Name:    "Index",
			Method:  "GET",
			Pattern: "/",
			Handler: h.IndexHandler(conf),
		},
		types.Route{
			Name:    "Version",
			Method:  "GET",
			Pattern: "/version",
			Handler: h.Version(conf),
		},
		types.Route{
			Name:    "Metrics",
			Method:  "GET",
			Pattern: "/metrics",
			Handler: prometheus.Handler(),
		},
	}
}
