package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/types"
)

var (
	defaultFilter = &mesos.Filters{RefuseSeconds: proto.Float64(10)}
	maxRetries    = 5

	// ErrQueueFull is returned in the event of a full queue. This allows the caller
	// to handle this as they see fit.
	ErrQueueFull = errors.New("task queue is full")
)

// eremeticScheduler holds the structure of the Eremetic Scheduler
type eremeticScheduler struct {
	tasksCreated int
	initialised  bool

	// task to start
	tasks chan string

	// This channel is closed when the program receives an interrupt,
	// signalling that the program should shut down.
	shutdown chan struct{}

	// Handle for current reconciliation job
	reconcile *Reconcile

	// Handler for storing tasks
	database database.TaskDB
}

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

func Create(settings *Settings, db database.TaskDB) *eremeticScheduler {
	return &eremeticScheduler{
		shutdown: make(chan struct{}),
		tasks:    make(chan string, settings.MaxQueueSize),
		database: db,
	}
}

func (s *eremeticScheduler) Reconcile(driver sched.SchedulerDriver) {
	if s.reconcile != nil {
		s.reconcile.Cancel()
	}
	s.reconcile = ReconcileTasks(driver, s.database)
}

func (s *eremeticScheduler) newTask(spec types.EremeticTask, offer *mesos.Offer) (types.EremeticTask, *mesos.TaskInfo) {
	return createTaskInfo(spec, offer)
}

// Registered is called when the Scheduler is Registered
func (s *eremeticScheduler) Registered(driver sched.SchedulerDriver, frameworkID *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	logrus.WithFields(logrus.Fields{
		"framework_id": frameworkID.GetValue(),
		"master_id":    masterInfo.GetId(),
		"master":       masterInfo.GetHostname(),
	}).Debug("Framework registered with master.")

	if !s.initialised {
		driver.ReconcileTasks([]*mesos.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Reregistered is called when the Scheduler is Reregistered
func (s *eremeticScheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	logrus.WithFields(logrus.Fields{
		"master_id": masterInfo.GetId(),
		"master":    masterInfo.GetHostname(),
	}).Debug("Framework re-registered with master.")
	if !s.initialised {
		driver.ReconcileTasks([]*mesos.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Disconnected is called when the Scheduler is Disconnected
func (s *eremeticScheduler) Disconnected(sched.SchedulerDriver) {
	logrus.Debugf("Framework disconnected with master")
}

// ResourceOffers handles the Resource Offers
func (s *eremeticScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
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
			offer, offers = matchOffer(t, offers)

			if offer == nil {
				logrus.WithField("task_id", tid).Warn("Unable to find a matching offer")
				TasksDelayed.Inc()
				go func() { s.tasks <- tid }()
				break loop
			}

			logrus.WithFields(logrus.Fields{
				"task_id":  tid,
				"offer_id": offer.Id.GetValue(),
			}).Debug("Preparing to launch task")

			t, task := s.newTask(t, offer)
			s.database.PutTask(&t)
			driver.LaunchTasks([]*mesos.OfferID{offer.Id}, []*mesos.TaskInfo{task}, defaultFilter)
			TasksLaunched.Inc()
			QueueSize.Dec()

			continue
		default:
			break loop
		}
	}

	logrus.Debug("No tasks to launch. Declining offers.")
	for _, offer := range offers {
		driver.DeclineOffer(offer.Id, defaultFilter)
	}
}

// StatusUpdate takes care of updating the status
func (s *eremeticScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	id := status.TaskId.GetValue()

	logrus.WithFields(logrus.Fields{
		"task_id": id,
		"status":  status.State.String(),
	}).Debug("Received task status update")

	task, err := s.database.ReadUnmaskedTask(id)
	if err != nil {
		logrus.WithError(err).WithField("task_id", id).Debug("Unable to read task from database")
	}

	if task.ID == "" {
		task = types.EremeticTask{
			ID:      id,
			SlaveId: status.SlaveId.GetValue(),
		}
	}

	if *status.State == mesos.TaskState_TASK_RUNNING && !task.IsRunning() {
		TasksRunning.Inc()
	}

	var shouldRetry bool
	if *status.State == mesos.TaskState_TASK_FAILED && !task.WasRunning() {
		if task.Retry >= maxRetries {
			logrus.WithFields(logrus.Fields{
				"task_id": id,
				"retries": task.Retry,
			}).Warn("Giving up on launching task")
		} else {
			shouldRetry = true
		}
	}

	if types.IsTerminal(status.State) {
		var seq string
		if shouldRetry {
			seq = "retry"
		} else {
			seq = "final"
		}
		TasksTerminated.With(prometheus.Labels{
			"status":   status.State.String(),
			"sequence": seq,
		}).Inc()
		if task.WasRunning() {
			TasksRunning.Dec()
		}
	}

	task.UpdateStatus(types.Status{
		Status: status.State.String(),
		Time:   time.Now().Unix(),
	})

	if shouldRetry {
		logrus.WithField("task_id", id).Info("Re-scheduling task that never ran.")
		task.UpdateStatus(types.Status{
			Status: mesos.TaskState_TASK_STAGING.String(),
			Time:   time.Now().Unix(),
		})
		task.Retry++
		go func() {
			QueueSize.Inc()
			s.tasks <- id
		}()
	} else if types.IsTerminal(status.State) {
		NotifyCallback(&task)
	}

	s.database.PutTask(&task)
}

func (s *eremeticScheduler) FrameworkMessage(
	driver sched.SchedulerDriver,
	executorID *mesos.ExecutorID,
	slaveID *mesos.SlaveID,
	message string) {

	logrus.Debug("Getting a framework message")
	switch executorID.GetValue() {
	case "eremetic-executor":
		var result interface{}
		err := json.Unmarshal([]byte(message), &result)
		if err != nil {
			logrus.WithError(err).Error("Unable to unmarshal result")
			return
		}
		logrus.Debug(message)

	default:
		logrus.WithField("executor_id", executorID.GetValue()).Debug("Received a message from an unknown executor.")
	}
}

func (s *eremeticScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	logrus.WithField("offer_id", offerID).Debug("Offer Rescinded")
}
func (s *eremeticScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	logrus.WithField("slave_id", slaveID).Debug("Slave lost")
}
func (s *eremeticScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	logrus.WithFields(logrus.Fields{
		"slave_id":    slaveID,
		"executor_id": executorID,
	}).Debug("Executor on slave was lost")
}

func (s *eremeticScheduler) Error(_ sched.SchedulerDriver, err string) {
	logrus.WithError(errors.New(err)).Debug("Received an error")
}

func nextID(s *eremeticScheduler) int {
	id := s.tasksCreated
	s.tasksCreated++
	return id
}

func (s *eremeticScheduler) ScheduleTask(request types.Request) (string, error) {
	logrus.WithFields(logrus.Fields{
		"docker_image": request.DockerImage,
		"command":      request.Command,
	}).Debug("Adding task to queue")

	task, err := types.NewEremeticTask(request, fmt.Sprintf("Eremetic task %d", nextID(s)))
	if err != nil {
		return "", err
	}

	select {
	case s.tasks <- task.ID:
		s.database.PutTask(&task)
		TasksCreated.Inc()
		QueueSize.Inc()
		return task.ID, nil
	case <-time.After(time.Duration(1) * time.Second):
		return "", ErrQueueFull
	}
}

func (s *eremeticScheduler) Stop() {
	close(s.shutdown)
}
