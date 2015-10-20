package types

import "net/http"

// Route enforces the structure of a route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a collection of route structs
type Routes []Route
