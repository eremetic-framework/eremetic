package routes

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

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
		HandlerFunc: AddTask,
	},
}

// AddTask adds a new task
func AddTask(w http.ResponseWriter, r *http.Request) {
	var request types.Request

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	handleError(err, w)

	err = json.Unmarshal(body, &request)
	handleError(err, w)

	createRequest(request, w)
}

func createRequest(request types.Request, w http.ResponseWriter) {
	writeJSON(http.StatusOK, request, w)
}

func writeJSON(status int, data interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleError(err error, w http.ResponseWriter) {
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
}
