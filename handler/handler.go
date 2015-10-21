package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/alde/eremetic/types"
	"github.com/alde/eremetic/zook"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/spf13/viper"
)

// CreateRequest handles creating a request for resources
func CreateRequest(request types.Request, w http.ResponseWriter) {
	WriteJSON(202, nil, w)
	go handle(request)
}

// WriteJSON handles writing a JSON response back to the HTTP socket
func WriteJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
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
		log.Printf("Unable to create scheduler driver: %s", err)
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

		log.Println("Eremetic is shutting down")

		// we have shut down
		driver.Stop(false)
	}()

	if status, err := driver.Run(); err != nil {
		log.Printf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
	}
	log.Println("Exiting...")
}
