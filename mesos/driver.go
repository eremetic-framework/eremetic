package mesos

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/auth"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/backoff"
	"github.com/mesos/mesos-go/api/v1/lib/encoding"
	"github.com/mesos/mesos-go/api/v1/lib/extras/scheduler/controller"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpsched"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler/calls"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler/events"
	"golang.org/x/net/context"
)

var (
	RegistrationMinBackoff = 1 * time.Second
	RegistrationMaxBackoff = 15 * time.Second
)

type Driver struct {
	settings   *Settings
	scheduler  *Scheduler
	caller     calls.Caller
	credential *mesosproto.Credential
	controller controller.Controller
}

func getFrameworkID(scheduler *Scheduler) *mesos.FrameworkID {
	if scheduler.frameworkID != "" {
		return &mesos.FrameworkID{
			Value: scheduler.frameworkID,
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

func createCaller(settings *Settings) calls.Caller {
	return httpsched.NewCaller(
		httpcli.New(
			httpcli.Endpoint(settings.Master),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					//FIXME: authConfigOpt,
					httpcli.Timeout(20*time.Second),
					httpcli.Transport(func(t *http.Transport) {
						// all calls should be ack'd by the server within this interval.
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2 // don't depend on go's default
					}),
				),
			),
		),
	)
}

func createEventsMux(sched *Scheduler) *events.Mux {
	return events.NewMux(
		events.DefaultHandler(events.HandlerFunc(controller.DefaultHandler)),
		events.MapFuncs(map[scheduler.Event_Type]events.HandlerFunc{
			scheduler.Event_FAILURE: func(e *scheduler.Event) error {
				f := e.GetFailure()
				logrus.WithFields(logrus.Fields{
					"executor_id": f.ExecutorID,
					"agent_id":    f.AgentID,
					"status":      f.Status,
				}).Debug("Got failure event")
				return nil
			},
			scheduler.Event_OFFERS: func(e *scheduler.Event) error {
				offers := e.GetOffers()
				logrus.WithFields(logrus.Fields{
					"num_offers": len(offers.GetOffers()),
				}).Debug("Got offers event")
				sched.ResourceOffers(
					offers.GetOffers(),
				)
				return nil
			},
			scheduler.Event_INVERSE_OFFERS: func(e *scheduler.Event) error {
				offers := e.GetInverseOffers()
				logrus.WithFields(logrus.Fields{
					"num_offers": len(offers.GetInverseOffers()),
				}).Debug("Got offers event")
				return nil
			},
			scheduler.Event_UPDATE: func(e *scheduler.Event) error {
				// FIXME: Should ACK
				update := e.GetUpdate()
				status := update.GetStatus()
				logrus.WithFields(logrus.Fields{
					"task_id": status.TaskID.Value,
					"state":   status.GetState().String(),
					"uuid":    string(status.GetUUID()),
				}).Debug("Got update event")
				sched.StatusUpdate(
					update.GetStatus(),
				)
				if len(status.GetUUID()) > 0 {
					ack := calls.Acknowledge(
						status.GetAgentID().GetValue(),
						status.TaskID.Value,
						status.GetUUID(),
					)
					if err := calls.CallNoData(sched.caller, ack); err != nil {
						logrus.WithError(err).Warn("Failed to ack status update")
					}
				}
				return nil
			},
			scheduler.Event_SUBSCRIBED: func(e *scheduler.Event) error {
				subscribed := e.GetSubscribed()
				logrus.WithFields(logrus.Fields{
					"framework_id": subscribed.GetFrameworkID().GetValue(),
				}).Debug("Got subscribed event")
				sched.Subscribed(
					subscribed.GetFrameworkID(),
				)
				return nil
			},
			scheduler.Event_MESSAGE: func(e *scheduler.Event) error {
				message := e.GetMessage()
				logrus.WithFields(logrus.Fields{
					"agent_id": message.GetAgentID().Value,
					"executor": message.GetExecutorID().Value,
				}).Debug("Got message event")
				return nil
			},
			scheduler.Event_RESCIND: func(e *scheduler.Event) error {
				rescind := e.GetRescind()
				logrus.WithFields(logrus.Fields{
					"offer_id": rescind.GetOfferID().Value,
				}).Debug("Got rescind event")
				return nil
			},
			scheduler.Event_RESCIND_INVERSE_OFFER: func(e *scheduler.Event) error {
				rescind := e.GetRescindInverseOffer()
				logrus.WithFields(logrus.Fields{
					"offer_id": rescind.GetInverseOfferID().Value,
				}).Debug("Got rescind inverse offer event")
				return nil
			},
		}),
	)
}

func createDriver(sched *Scheduler, settings *Settings) (*Driver, error) {
	credential, err := getCredential(settings)
	if err != nil {
		return nil, err
	}

	if settings.Master == "" {
		return nil, errors.New("Missing master location URL.")
	}

	return &Driver{
		settings:   settings,
		scheduler:  sched,
		caller:     sched.caller,
		credential: credential,
		controller: controller.New(),
	}, nil
}

func (d *Driver) Run(shutdown chan struct{}) error {
	return d.controller.Run(controller.Config{
		Context: &controller.ContextAdapter{
			FrameworkIDFunc: func() string { return d.scheduler.frameworkID },
			DoneFunc: func() bool {
				select {
				case <-shutdown:
					return true
				default:
					return false
				}
			},
			ErrorFunc: func(err error) {
				if err != nil {
					logrus.WithError(err).Error("Framework error")
				}
				logrus.Info("disconnected")
			},
		},
		Framework: &mesos.FrameworkInfo{
			ID:              getFrameworkID(d.scheduler),
			Name:            d.settings.Name,
			User:            d.settings.User,
			Checkpoint:      &d.settings.Checkpoint,
			FailoverTimeout: &d.settings.FailoverTimeout,
			Principal:       getPrincipalID(d.credential),
		},
		Caller:             d.caller,
		Handler:            createEventsMux(d.scheduler),
		RegistrationTokens: backoff.Notifier(RegistrationMinBackoff, RegistrationMaxBackoff, shutdown),
	})
}
