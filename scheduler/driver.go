package scheduler

import (
	"net"

	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/spf13/viper"
)

func getFrameworkID() *mesos.FrameworkID {
	id := viper.GetString("framework_id")
	if id != "" {
		return &mesos.FrameworkID{
			Value: proto.String(id),
		}
	}
	return nil
}

func createDriver(scheduler *eremeticScheduler) (*sched.MesosSchedulerDriver, error) {
	publishedAddr := net.ParseIP(viper.GetString("messenger_address"))
	bindingPort := uint16(viper.GetInt("messenger_port"))

	return sched.NewMesosSchedulerDriver(sched.DriverConfig{
		Master: viper.GetString("master"),
		Framework: &mesos.FrameworkInfo{
			Id:              getFrameworkID(),
			Name:            proto.String(viper.GetString("name")),
			User:            proto.String(viper.GetString("user")),
			Checkpoint:      proto.Bool(viper.GetBool("checkpoint")),
			FailoverTimeout: proto.Float64(viper.GetFloat64("failover_timeout")),
		},
		Scheduler:        scheduler,
		BindingAddress:   net.ParseIP("0.0.0.0"),
		PublishedAddress: publishedAddr,
		BindingPort:      bindingPort,
	})
}
