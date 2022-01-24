package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rockerbox/eremetic"
	"github.com/rockerbox/eremetic/config"
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

	return router
}

func routes(h Handler, conf *config.Config) Routes {
	v0routes := apiV0Routes(h, conf)
	v1routes := apiV1Routes(h, conf)
	apiRoutes := append(v0routes, v1routes...)
	return append(Routes{
		Route{
			Name:    "Index",
			Method:  "GET",
			Pattern: "/",
			Handler: h.IndexHandler(conf),
		},
		Route{
			Name:    "Metrics",
			Method:  "GET",
			Pattern: "/metrics",
			Handler: prometheus.Handler(),
		},
	}, apiRoutes...)
}
