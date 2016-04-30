package scheduler

import (
	"io/ioutil"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
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

func getPrincipalID(credential *mesos.Credential) *string {
	if credential != nil {
		return credential.Principal
	}
	return nil
}

func getCredential() (*mesos.Credential, error) {
	if viper.IsSet("credential_file") {
		content, err := ioutil.ReadFile(viper.GetString("credential_file"))
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"credential_file": viper.GetString("credential_file"),
			}).Error("Unable to read credential_file")
			return nil, err
		}
		fields := strings.Fields(string(content))

		if len(fields) != 2 {
			err := errors.New("Unable to parse credentials")
			logrus.WithError(err).WithFields(logrus.Fields{
				"credential_file": viper.GetString("credential_file"),
			}).Error("Should only contain a key and a secret separated by whitespace")
			return nil, err
		}

		logrus.WithField("principal", fields[0]).Info("Successfully loaded principal")
		return &mesos.Credential{
			Principal: proto.String(fields[0]),
			Secret:    proto.String(fields[1]),
		}, nil
	}
	logrus.Debug("No credentials specified in configuration")
	return nil, nil
}

func getAuthContext(ctx context.Context) context.Context {
	return auth.WithLoginProvider(ctx, "SASL")
}

func createDriver(scheduler *eremeticScheduler) (*sched.MesosSchedulerDriver, error) {
	publishedAddr := net.ParseIP(viper.GetString("messenger_address"))
	bindingPort := uint16(viper.GetInt("messenger_port"))
	credential, err := getCredential()

	if err != nil {
		return nil, err
	}

	return sched.NewMesosSchedulerDriver(sched.DriverConfig{
		Master: viper.GetString("master"),
		Framework: &mesos.FrameworkInfo{
			Id:              getFrameworkID(),
			Name:            proto.String(viper.GetString("name")),
			User:            proto.String(viper.GetString("user")),
			Checkpoint:      proto.Bool(viper.GetBool("checkpoint")),
			FailoverTimeout: proto.Float64(viper.GetFloat64("failover_timeout")),
			Principal:       getPrincipalID(credential),
		},
		Scheduler:        scheduler,
		BindingAddress:   net.ParseIP("0.0.0.0"),
		PublishedAddress: publishedAddr,
		BindingPort:      bindingPort,
		Credential:       credential,
		WithAuthContext:  getAuthContext,
	})
}
