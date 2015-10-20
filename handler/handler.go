package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alde/eremetic/types"
)

// CreateRequest handles creating a request for resources
func CreateRequest(request types.Request, w http.ResponseWriter) {
	WriteJSON(http.StatusOK, request, w)
}

// WriteJSON handles writing a JSON response back to the HTTP socket
func WriteJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
