package routes

import (
	"github.com/alde/eremetic/handler"
	"github.com/alde/eremetic/types"
	"github.com/gorilla/mux"
)

// Create is used to create a new router
func Create() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

var routes = types.Routes{
	types.Route{
		Name:        "AddTask",
		Method:      "POST",
		Pattern:     "/task",
		HandlerFunc: handler.AddTask,
	},
	types.Route{
		Name:        "Status",
		Method:      "GET",
		Pattern:     "/task/{taskId}",
		HandlerFunc: handler.GetTaskInfo,
	},
}
