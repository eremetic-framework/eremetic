package handler

import (
	"encoding/json"
	"net/http"

	log "github.com/dmuth/google-go-log4go"

	"github.com/alde/eremetic/types"
)

var requests = make(chan *types.Request)

// CreateRequest handles creating a request for resources
func CreateRequest(request types.Request, w http.ResponseWriter) {
	defer WriteJSON(202, nil, w)
	log.Debugf("Adding request for '%s' to queue.", request.DockerImage)
	requests <- &request
}

// WriteJSON handles writing a JSON response back to the HTTP socket
func WriteJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// Run the RequestChannel Listener
func Run() {
	scheduler := createEremeticScheduler()
	driver, err := createDriver(scheduler)

	if err != nil {
		log.Errorf("Unable to create scheduler driver: %s", err)
		return
	}

	defer close(scheduler.shutdown)
	defer driver.Stop(false)

	go func() {
		if status, err := driver.Run(); err != nil {
			log.Errorf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
		}
		log.Info("Exiting...")
	}()

	log.Debug("Entering handler.Run loop")
	for {
		select {
		case req := <-requests:
			log.Debug("Found a request in the queue!")
			scheduleTasks(scheduler, *req)
		}
	}
}
