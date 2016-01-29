package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/handler"
	"github.com/klarna/eremetic/types"
)

var (
	defaultFilter = &mesos.Filters{RefuseSeconds: proto.Float64(10)}
	maxRetries    = 5
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
	// This channel is closed after shutdown is closed, and only when all
	// outstanding tasks have been cleaned up
	done chan struct{}

	// Handle for current reconciliation job
	reconcile *Reconcile
}

func (s *eremeticScheduler) Reconcile(driver sched.SchedulerDriver) {
	if s.reconcile != nil {
		s.reconcile.Cancel()
	}
	s.reconcile = ReconcileTasks(driver)
}

func (s *eremeticScheduler) newTask(spec types.EremeticTask, offer *mesos.Offer) (types.EremeticTask, *mesos.TaskInfo) {
	return createTaskInfo(spec, offer)
}

// Registered is called when the Scheduler is Registered
func (s *eremeticScheduler) Registered(driver sched.SchedulerDriver, frameworkID *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	logrus.WithFields(logrus.Fields{
		"framework": frameworkID.GetValue(),
		"master":    masterInfo.GetHostname(),
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
	logrus.WithField("master", masterInfo).Debug("Framework re-registered with master.")
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
			logrus.WithField("task", tid).Debug("Trying to find offer to launch task with")
			t, _ := database.ReadTask(tid)
			offer, offers = matchOffer(t, offers)

			if offer == nil {
				logrus.WithField("task", tid).Warn("Unable to find a matching offer")
				TasksDelayed.Inc()
				go func() { s.tasks <- tid }()
				break loop
			}

			logrus.WithFields(logrus.Fields{
				"task":  tid,
				"offer": offer.Id.GetValue(),
			}).Debug("Preparing to launch task")

			t, task := s.newTask(t, offer)
			database.PutTask(&t)
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
		"task":   id,
		"status": status.State.String(),
	}).Debug("Received task status update")

	task, err := database.ReadTask(id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"task": id,
		}).Debug("Unable to read task from database")
	}

	if task.ID == "" {
		task = types.EremeticTask{
			ID:      id,
			SlaveId: status.SlaveId.GetValue(),
		}
	}

	if !task.IsRunning() && *status.State == mesos.TaskState_TASK_RUNNING {
		TasksRunning.Inc()
	}

	if types.IsTerminal(status.State) {
		TasksTerminated.With(prometheus.Labels{"status": status.State.String()}).Inc()
		if task.WasRunning() {
			TasksRunning.Dec()
		}
	}

	task.UpdateStatus(types.Status{
		Status: status.State.String(),
		Time:   time.Now().Unix(),
	})

	if *status.State == mesos.TaskState_TASK_FAILED && !task.WasRunning() {
		if task.Retry >= maxRetries {
			logrus.WithFields(logrus.Fields{
				"task":    id,
				"retries": task.Retry,
			}).Warn("Giving up on launching task")
		} else {
			logrus.WithField("task", id).Info("Re-scheduling task that never ran.")
			task.UpdateStatus(types.Status{
				Status: mesos.TaskState_TASK_STAGING.String(),
				Time:   time.Now().Unix(),
			})
			task.Retry++
			go func() {
				QueueSize.Inc()
				s.tasks <- id
			}()
		}
	}

	if types.IsTerminal(status.State) {
		handler.NotifyCallback(&task)
	}

	database.PutTask(&task)
}

func (s *eremeticScheduler) FrameworkMessage(
	driver sched.SchedulerDriver,
	executorID *mesos.ExecutorID,
	slaveID *mesos.SlaveID,
	message string) {

	logrus.Debug("Getting a framework message")
	switch *executorID.Value {
	case "eremetic-executor":
		var result interface{}
		err := json.Unmarshal([]byte(message), &result)
		if err != nil {
			logrus.WithError(err).Error("Unable to unmarshal result")
			return
		}
		logrus.Debug(message)

	default:
		logrus.WithField("framework", executorID.GetValue()).Debug("Received a message from an unknown framework.")
	}
}

func (s *eremeticScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	logrus.WithField("offer", offerID).Debug("Offer Rescinded")
}
func (s *eremeticScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	logrus.WithField("slave", slaveID).Debug("Slave lost")
}
func (s *eremeticScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	logrus.WithFields(logrus.Fields{
		"slave":    slaveID,
		"executor": executorID,
	}).Debug("Executor on slave was lost")
}

func (s *eremeticScheduler) Error(_ sched.SchedulerDriver, err string) {
	logrus.WithField("err", err).Debug("Received an error")
}

func createEremeticScheduler() *eremeticScheduler {
	s := &eremeticScheduler{
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
		tasks:    make(chan string, 100),
	}
	return s
}

func nextID(s *eremeticScheduler) int {
	id := s.tasksCreated
	s.tasksCreated++
	return id
}

func (s *eremeticScheduler) ScheduleTask(request types.Request) (string, error) {
	logrus.WithField("docker_image", request.DockerImage).Debug("Adding task to queue")

	request.Name = fmt.Sprintf("Eremetic task %d", nextID(s))

	task, err := createEremeticTask(request)
	if err != nil {
		logrus.Error(err.Error())
		return "", err
	}

	TasksCreated.Inc()
	QueueSize.Inc()
	database.PutTask(&task)
	s.tasks <- task.ID
	return task.ID, nil
}
