package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	mesossched "github.com/mesos/mesos-go/api/v0/scheduler"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cybricio/eremetic"
	"github.com/cybricio/eremetic/metrics"
)

var (
	defaultFilter = &mesosproto.Filters{RefuseSeconds: proto.Float64(10)}
	maxRetries    = 5
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

	frameworkID  string
	tasksCreated int
	initialised  bool
	driver       mesossched.SchedulerDriver

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
		frameworkID: settings.FrameworkID,
	}
}

// Run the eremetic scheduler
func (s *Scheduler) Run() {
	driver, err := createDriver(s, s.settings)
	s.driver = driver

	if err != nil {
		logrus.WithError(err).Error("Unable to create scheduler driver")
	}

	go func() {
		<-s.shutdown
		driver.Stop(false)
	}()

	if status, err := driver.Run(); err != nil {
		logrus.WithError(err).WithField("status", status.String()).Error("Framework stopped")
	}

	logrus.Info("Exiting...")
}

// Reconcile reconciles the currently scheduled tasks.
func (s *Scheduler) Reconcile(driver mesossched.SchedulerDriver) {
	if s.reconcile != nil {
		s.reconcile.Cancel()
	}
	s.reconcile = reconcileTasks(driver, s.database)
}

// Registered is called when the Scheduler is Registered
func (s *Scheduler) Registered(driver mesossched.SchedulerDriver, frameworkID *mesosproto.FrameworkID, masterInfo *mesosproto.MasterInfo) {
	logrus.WithFields(logrus.Fields{
		"framework_id": frameworkID.GetValue(),
		"master_id":    masterInfo.GetId(),
		"master":       masterInfo.GetHostname(),
	}).Debug("Framework registered with master.")

	s.frameworkID = frameworkID.GetValue()
	if !s.initialised {
		driver.ReconcileTasks([]*mesosproto.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Reregistered is called when the Scheduler is Reregistered
func (s *Scheduler) Reregistered(driver mesossched.SchedulerDriver, masterInfo *mesosproto.MasterInfo) {
	logrus.WithFields(logrus.Fields{
		"master_id": masterInfo.GetId(),
		"master":    masterInfo.GetHostname(),
	}).Debug("Framework re-registered with master.")
	if !s.initialised {
		driver.ReconcileTasks([]*mesosproto.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Disconnected is called when the Scheduler is Disconnected
func (s *Scheduler) Disconnected(mesossched.SchedulerDriver) {
	logrus.Debugf("Framework disconnected with master")
}

// ResourceOffers handles the Resource Offers
func (s *Scheduler) ResourceOffers(driver mesossched.SchedulerDriver, offers []*mesosproto.Offer) {
	logrus.WithField("offers", len(offers)).Debug("Received offers")
	var offer *mesosproto.Offer

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
				"task_id":  task.TaskId.GetValue(),
				"offer_id": offer.Id.GetValue(),
			}).Debug("Preparing to launch task")
			t.UpdateStatus(eremetic.Status{
				Status: eremetic.TaskStaging,
				Time:   time.Now().Unix(),
			})
			s.database.PutTask(&t)
			_, err := driver.LaunchTasks([]*mesosproto.OfferID{offer.Id}, []*mesosproto.TaskInfo{task}, defaultFilter)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"task_id":  task.TaskId.GetValue(),
					"offer_id": offer.Id.GetValue(),
				}).WithError(err).Warn("Failed to launch task")
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskError,
					Time:   time.Now().Unix(),
				})
			} else {
				metrics.TasksLaunched.Inc()
			}
			metrics.QueueSize.Dec()

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
func (s *Scheduler) StatusUpdate(driver mesossched.SchedulerDriver, status *mesosproto.TaskStatus) {
	id := status.TaskId.GetValue()
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
			AgentID: status.SlaveId.GetValue(),
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

// FrameworkMessage is invoked when an executor sends a message.
func (s *Scheduler) FrameworkMessage(
	driver mesossched.SchedulerDriver,
	executorID *mesosproto.ExecutorID,
	slaveID *mesosproto.SlaveID,
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

// OfferRescinded is invoked when an offer is no longer valid.
func (s *Scheduler) OfferRescinded(_ mesossched.SchedulerDriver, offerID *mesosproto.OfferID) {
	logrus.WithField("offer_id", offerID).Debug("Offer Rescinded")
}

// SlaveLost is invoked when a slave has been determined unreachable.
func (s *Scheduler) SlaveLost(_ mesossched.SchedulerDriver, slaveID *mesosproto.SlaveID) {
	logrus.WithField("slave_id", slaveID).Debug("Slave lost")
}

// ExecutorLost is invoked when an executor has exited/terminated.
func (s *Scheduler) ExecutorLost(_ mesossched.SchedulerDriver, executorID *mesosproto.ExecutorID, agentID *mesosproto.SlaveID, status int) {
	logrus.WithFields(logrus.Fields{
		"agent_id":    agentID,
		"executor_id": executorID,
	}).Debug("Executor on agent was lost")
}

// Error is invoked when there is an unrecoverable error in the scheduler or scheduler driver.
func (s *Scheduler) Error(_ mesossched.SchedulerDriver, err string) {
	logrus.WithError(errors.New(err)).Debug("Received an error")
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

	_, err = s.driver.KillTask(&mesosproto.TaskID{Value: proto.String(taskId)})
	return err
}

// Stop triggers a shutdown of the scheduler.
func (s *Scheduler) Stop() {
	close(s.shutdown)
}

func nextID(s *Scheduler) string {
	letters := []rune("bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	s.tasksCreated++

	return string(b)
}
