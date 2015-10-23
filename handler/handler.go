package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"

	log "github.com/dmuth/google-go-log4go"

	"github.com/alde/eremetic/types"
	"github.com/alde/eremetic/zook"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/spf13/viper"
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
	log.Debug("Entering handler.Run loop")
	for {
		select {
		case req := <-requests:
			log.Debug("Found a request in the queue!")
			handle(*req)
		}
	}
}

func handle(request types.Request) {
	publishedAddr := net.ParseIP(viper.GetString("messenger_address"))
	bindingPort := uint16(viper.GetInt("messenger_port"))
	master := zook.DiscoverMaster(viper.GetString("zookeeper"))
	scheduler := createEremeticScheduler(request)
	defer close(scheduler.shutdown)

	driver, err := sched.NewMesosSchedulerDriver(sched.DriverConfig{
		Master: master,
		Framework: &mesos.FrameworkInfo{
			Name: proto.String("Eremetic"),
			User: proto.String(""),
		},
		Scheduler:        scheduler,
		BindingAddress:   net.ParseIP("0.0.0.0"),
		PublishedAddress: publishedAddr,
		BindingPort:      bindingPort,
	})
	if err != nil {
		log.Errorf("Unable to create scheduler driver: %s", err)
		return
	}

	// Catch interrupt
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		if s != os.Interrupt {
			return
		}

		log.Info("Eremetic is shutting down")

		// we have shut down
		driver.Stop(false)
	}()

	if status, err := driver.Run(); err != nil {
		log.Errorf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
	}
	log.Info("Exiting...")
}
