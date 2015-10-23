package handler

import (
	"net"

	"github.com/alde/eremetic/zook"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/spf13/viper"
)

func createDriver(scheduler *eremeticScheduler) (*sched.MesosSchedulerDriver, error) {
	publishedAddr := net.ParseIP(viper.GetString("messenger_address"))
	bindingPort := uint16(viper.GetInt("messenger_port"))
	master := zook.DiscoverMaster(viper.GetString("zookeeper"))

	return sched.NewMesosSchedulerDriver(sched.DriverConfig{
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
}
