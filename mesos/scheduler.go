package mesos

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler/calls"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/metrics"
)

var (
	maxRetries = 5
)

// Settings holds configuration values for the scheduler
type Settings struct {
	MaxQueueSize     int
	Master           string
	FrameworkID      string
	CredentialFile   string
	Name             string
	User             string
	MessengerAddress string
	MessengerPort    uint16
	Checkpoint       bool
	FailoverTimeout  float64
}

// Scheduler holds the structure of the Eremetic Scheduler
type Scheduler struct {
	settings *Settings

	frameworkID string
	initialised bool
	random      *rand.Rand
	caller      calls.Caller

	// task to start
	tasks chan string

	// This channel is closed when the program receives an interrupt,
	// signalling that the program should shut down.
	shutdown chan struct{}

	// Handle for current reconciliation job
	reconcile *reconciler

	// Handler for storing tasks
	database eremetic.TaskDB
}

// NewScheduler returns a new instance of the default scheduler.
func NewScheduler(settings *Settings, db eremetic.TaskDB) *Scheduler {
	return &Scheduler{
		settings:    settings,
		shutdown:    make(chan struct{}),
		tasks:       make(chan string, settings.MaxQueueSize),
		database:    db,
		random:      rand.New(rand.NewSource(time.Now().Unix())),
		frameworkID: settings.FrameworkID,
	}
}

// Run the eremetic scheduler
func (s *Scheduler) Run() {
	s.caller = createCaller(s.settings)
	driver, err := createDriver(s, s.settings)
	if err != nil {
		logrus.WithError(err).Error("Unable to create scheduler driver")
	}
	if err := driver.Run(s.shutdown); err != nil {
		logrus.WithError(err).Error("Framework stopped")
	}
	logrus.Info("Exiting...")
}

// Reconcile reconciles the currently scheduled tasks.
func (s *Scheduler) Reconcile() {
	if s.reconcile != nil {
		s.reconcile.Cancel()
	}
	s.reconcile = s.reconcileTasks()
}

// Subscribed is called whne the Scheduler is Subscribed
func (s *Scheduler) Subscribed(frameworkID *mesos.FrameworkID) {
	logrus.WithFields(logrus.Fields{
		"framework_id": frameworkID.GetValue(),
	}).Debug("Framework subscribed to master.")
	s.frameworkID = frameworkID.GetValue()
	if !s.initialised {
		frameworkOpt := calls.Framework(s.frameworkID)
		reconcile := calls.Reconcile().With(frameworkOpt)
		if err := calls.CallNoData(s.caller, reconcile); err != nil {
			logrus.WithError(err).Warn("Failed to send reconciliation request")
		}
	} else {
		s.Reconcile()
	}
}

// ResourceOffers handles the Resource Offers
func (s *Scheduler) ResourceOffers(offers []mesos.Offer) {
	refuseOpt := calls.RefuseSecondsWithJitter(s.random, 10)
	frameworkOpt := calls.Framework(s.frameworkID)
	logrus.WithField("offers", len(offers)).Debug("Received offers")
	var offer *mesos.Offer

loop:
	for len(offers) > 0 {
		select {
		case <-s.shutdown:
			logrus.Info("Shutting down: declining offers")
			break loop
		case tid := <-s.tasks:
			logrus.WithField("task_id", tid).Debug("Trying to find offer to launch task with")
			t, _ := s.database.ReadUnmaskedTask(tid)

			if t.IsTerminating() {
				logrus.Debug("Dropping terminating task.")
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskKilled,
					Time:   time.Now().Unix(),
				})
				s.database.PutTask(&t)

				continue
			}
			offer, offers = matchOffer(t, offers)

			if offer == nil {
				logrus.WithField("task_id", tid).Warn("Unable to find a matching offer")
				metrics.TasksDelayed.Inc()
				go func() { s.tasks <- tid }()
				break loop
			}

			t, task := createTaskInfo(t, offer)
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID.GetValue(),
				"offer_id": offer.ID.GetValue(),
			}).Debug("Preparing to launch task")
			t.UpdateStatus(eremetic.Status{
				Status: eremetic.TaskStaging,
				Time:   time.Now().Unix(),
			})
			accept := calls.Accept(
				calls.OfferOperations{calls.OpLaunch(*task)}.WithOffers(offer.ID),
			).With(refuseOpt).With(frameworkOpt)
			if err := calls.CallNoData(s.caller, accept); err != nil {
				logrus.WithFields(logrus.Fields{
					"task_id":  task.TaskID.GetValue(),
					"offer_id": offer.ID.GetValue(),
				}).WithError(err).Warn("Failed to launch task")
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskError,
					Time:   time.Now().Unix(),
				})
			} else {
				metrics.TasksLaunched.Inc()
			}
			metrics.QueueSize.Dec()
			s.database.PutTask(&t)
			continue
		default:
			break loop
		}
	}

	logrus.Debug("No tasks to launch. Declining offers.")
	for _, offer := range offers {
		accept := calls.Accept(
			calls.OfferOperations{calls.OpLaunch()}.WithOffers(offer.ID),
		).With(refuseOpt).With(frameworkOpt)
		if err := calls.CallNoData(s.caller, accept); err != nil {
			logrus.WithField("offer_id", offer.ID).WithError(err).Info("Failed to decline offer")
		}
	}
}

// StatusUpdate takes care of updating the status
func (s *Scheduler) StatusUpdate(status mesos.TaskStatus) {
	id := status.TaskID.GetValue()
	newState := eremetic.TaskState(status.State.String())

	logrus.WithFields(logrus.Fields{
		"task_id": id,
		"status":  status.State.String(),
	}).Debug("Received task status update")

	task, err := s.database.ReadUnmaskedTask(id)
	if err != nil {
		logrus.WithError(err).WithField("task_id", id).Debug("Unable to read task from database")
	}

	if task.ID == "" {
		task = eremetic.Task{
			ID:      id,
			AgentID: status.AgentID.GetValue(),
		}
	}

	sandboxPath, err := extractSandboxPath(status.Data)
	if err != nil {
		logrus.WithError(err).Debug("Unable to extract sandbox path")
	}

	if sandboxPath != "" {
		task.SandboxPath = sandboxPath
	}

	if newState == eremetic.TaskRunning && !task.IsRunning() {
		metrics.TasksRunning.Inc()
	}

	var shouldRetry bool
	if newState == eremetic.TaskFailed && !task.WasRunning() {
		if task.Retry >= maxRetries {
			logrus.WithFields(logrus.Fields{
				"task_id": id,
				"retries": task.Retry,
			}).Warn("Giving up on launching task")
		} else {
			shouldRetry = true
		}
	}

	if eremetic.IsTerminal(newState) {
		var seq string
		if shouldRetry {
			seq = "retry"
		} else {
			seq = "final"
		}
		metrics.TasksTerminated.With(prometheus.Labels{
			"status":   string(newState),
			"sequence": seq,
		}).Inc()
		if task.WasRunning() {
			metrics.TasksRunning.Dec()
		}
	}

	task.UpdateStatus(eremetic.Status{
		Status: newState,
		Time:   time.Now().Unix(),
	})

	if shouldRetry {
		logrus.WithField("task_id", id).Info("Re-scheduling task that never ran.")
		task.UpdateStatus(eremetic.Status{
			Status: eremetic.TaskQueued,
			Time:   time.Now().Unix(),
		})
		task.Retry++
		go func() {
			metrics.QueueSize.Inc()
			s.tasks <- id
		}()
	} else if eremetic.IsTerminal(newState) {
		eremetic.NotifyCallback(&task)
	}

	s.database.PutTask(&task)
}

// ScheduleTask tries to register a new task in the database to be scheduled.
// If the queue is full the task will be dropped.
func (s *Scheduler) ScheduleTask(request eremetic.Request) (string, error) {
	logrus.WithFields(logrus.Fields{
		"docker_image":      request.DockerImage,
		"command":           request.Command,
		"agent_constraints": request.AgentConstraints,
		"ports":             request.Ports,
	}).Debug("Adding task to queue")

	if request.Name == "" {
		request.Name = fmt.Sprintf("Eremetic task %s", nextID(s))
	}

	task, err := eremetic.NewTask(request)
	if err != nil {
		return "", err
	}

	select {
	case s.tasks <- task.ID:
		s.database.PutTask(&task)
		metrics.TasksCreated.Inc()
		metrics.QueueSize.Inc()
		return task.ID, nil
	case <-time.After(time.Duration(1) * time.Second):
		return "", eremetic.ErrQueueFull
	}
}

func (s *Scheduler) Kill(taskId string) error {
	task, err := s.database.ReadTask(taskId)
	if err != nil {
		return err
	}

	if task.IsTerminated() {
		return fmt.Errorf("You can not kill that which is already dead.")
	}

	waiting := task.IsWaiting()

	logrus.Debugf("Marking task for killing.")
	task.UpdateStatus(eremetic.Status{
		Status: eremetic.TaskTerminating,
		Time:   time.Now().Unix(),
	})
	s.database.PutTask(&task)

	if waiting {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"task_id":  task.ID,
		"agent_id": task.AgentID,
	}).Debug("sending kill request")
	frameworkOpt := calls.Framework(s.frameworkID)
	kill := calls.Kill(task.ID, task.AgentID).With(frameworkOpt)
	err = calls.CallNoData(s.caller, kill)
	return err
}

// Stop triggers a shutdown of the scheduler.
func (s *Scheduler) Stop() {
	close(s.shutdown)
}

func nextID(_ *Scheduler) string {
	letters := []rune("bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
