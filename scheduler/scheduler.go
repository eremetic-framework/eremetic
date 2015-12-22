package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/dmuth/google-go-log4go"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/klarna/eremetic/database"
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
	log.Debugf("Framework %s registered with master %s", frameworkID.GetValue(), masterInfo.GetHostname())
	if !s.initialised {
		driver.ReconcileTasks([]*mesos.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Reregistered is called when the Scheduler is Reregistered
func (s *eremeticScheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Debugf("Framework re-registered with master %s", masterInfo)
	if !s.initialised {
		driver.ReconcileTasks([]*mesos.TaskStatus{})
		s.initialised = true
	} else {
		s.Reconcile(driver)
	}
}

// Disconnected is called when the Scheduler is Disconnected
func (s *eremeticScheduler) Disconnected(sched.SchedulerDriver) {
	log.Debugf("Framework disconnected with master")
}

// ResourceOffers handles the Resource Offers
func (s *eremeticScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log.Tracef("Received %d resource offers", len(offers))
	var offer *mesos.Offer

loop:
	for len(offers) > 0 {
		select {
		case <-s.shutdown:
			log.Info("Shutting down: declining offers")
			break loop
		case tid := <-s.tasks:
			log.Debugf("Trying to find offer to launch %s with", tid)
			t, _ := database.ReadTask(tid)
			offer, offers = matchOffer(t, offers)

			if offer == nil {
				log.Warnf("Could not find a matching offer for %s", tid)
				TasksDelayed.Inc()
				go func() { s.tasks <- tid }()
				break loop
			}

			log.Debugf("Preparing to launch task %s with offer %s", tid, offer.Id.GetValue())
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

	log.Trace("No tasks to launch. Declining offers.")
	for _, offer := range offers {
		driver.DeclineOffer(offer.Id, defaultFilter)
	}
}

// StatusUpdate takes care of updating the status
func (s *eremeticScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	id := status.TaskId.GetValue()

	log.Debugf("Received task status [%s] for task [%s]", status.State.String(), id)

	task, err := database.ReadTask(id)
	if err != nil {
		log.Debugf("Error reading task from database: %s", err)
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
		TasksRunning.Dec()
	}

	task.UpdateStatus(types.Status{
		Status: status.State.String(),
		Time:   time.Now().Unix(),
	})

	if *status.State == mesos.TaskState_TASK_FAILED && !task.WasRunning() {
		if task.Retry >= maxRetries {
			log.Warnf("giving up on %s after %d retry attempts", id, task.Retry)
		} else {
			log.Infof("task %s was never running. re-scheduling", id)
			task.UpdateStatus(types.Status{
				Status: mesos.TaskState_TASK_STAGING.String(),
				Time:   time.Now().Unix(),
			})
			task.Retry += 1
			go func() {
				QueueSize.Inc()
				s.tasks <- id
			}()
		}
	}

	database.PutTask(&task)
}

func (s *eremeticScheduler) FrameworkMessage(
	driver sched.SchedulerDriver,
	executorID *mesos.ExecutorID,
	slaveID *mesos.SlaveID,
	message string) {

	log.Debug("Getting a framework message")
	switch *executorID.Value {
	case "eremetic-executor":
		var result interface{}
		err := json.Unmarshal([]byte(message), &result)
		if err != nil {
			log.Errorf("Error deserializing Result: [%s]", err)
			return
		}
		log.Debug(message)

	default:
		log.Debugf("Received a framework message from some unknown source: %s", *executorID.Value)
	}
}

func (s *eremeticScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	log.Debugf("Offer %s rescinded", offerID)
}
func (s *eremeticScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	log.Debugf("Slave %s lost", slaveID)
}
func (s *eremeticScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	log.Debugf("Executor %s on slave %s was lost", executorID, slaveID)
}

func (s *eremeticScheduler) Error(_ sched.SchedulerDriver, err string) {
	log.Debugf("Receiving an error: %s", err)
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
	log.Debugf(
		"Adding task running on %s to queue",
		request.DockerImage)

	request.Name = fmt.Sprintf("Eremetic task %d", nextID(s))

	task, err := createEremeticTask(request)
	if err != nil {
		return "", err
	}

	TasksCreated.Inc()
	QueueSize.Inc()
	database.PutTask(&task)
	s.tasks <- task.ID
	return task.ID, nil
}
