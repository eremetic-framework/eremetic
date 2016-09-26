package mesos

import (
	"errors"
	"io/ioutil"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/mesosproto"
	mesossched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
)

func getFrameworkID(settings *Settings) *mesosproto.FrameworkID {
	if settings.FrameworkID != "" {
		return &mesosproto.FrameworkID{
			Value: proto.String(settings.FrameworkID),
		}
	}
	return nil
}

func getPrincipalID(credential *mesosproto.Credential) *string {
	if credential != nil {
		return credential.Principal
	}
	return nil
}

func getCredential(settings *Settings) (*mesosproto.Credential, error) {
	if settings.CredentialFile != "" {
		content, err := ioutil.ReadFile(settings.CredentialFile)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"credential_file": settings.CredentialFile,
			}).Error("Unable to read credential_file")
			return nil, err
		}
		fields := strings.Fields(string(content))

		if len(fields) != 2 {
			err := errors.New("Unable to parse credentials")
			logrus.WithError(err).WithFields(logrus.Fields{
				"credential_file": settings.CredentialFile,
			}).Error("Should only contain a key and a secret separated by whitespace")
			return nil, err
		}

		logrus.WithField("principal", fields[0]).Info("Successfully loaded principal")
		return &mesosproto.Credential{
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

func createDriver(scheduler *Scheduler, settings *Settings) (*mesossched.MesosSchedulerDriver, error) {
	publishedAddr := net.ParseIP(settings.MessengerAddress)
	bindingPort := settings.MessengerPort
	credential, err := getCredential(settings)

	if err != nil {
		return nil, err
	}

	return mesossched.NewMesosSchedulerDriver(mesossched.DriverConfig{
		Master: settings.Master,
		Framework: &mesosproto.FrameworkInfo{
			Id:              getFrameworkID(settings),
			Name:            proto.String(settings.Name),
			User:            proto.String(settings.User),
			Checkpoint:      proto.Bool(settings.Checkpoint),
			FailoverTimeout: proto.Float64(settings.FailoverTimeout),
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
