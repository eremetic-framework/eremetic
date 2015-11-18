package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alde/eremetic/database"
	"github.com/alde/eremetic/types"
	log "github.com/dmuth/google-go-log4go"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
)

var (
	defaultFilter = &mesos.Filters{RefuseSeconds: proto.Float64(1)}
)

// eremeticScheduler holds the structure of the Eremetic Scheduler
type eremeticScheduler struct {
	tasksCreated int

	// task to start
	tasks chan string

	// This channel is closed when the program receives an interrupt,
	// signalling that the program should shut down.
	shutdown chan struct{}
	// This channel is closed after shutdown is closed, and only when all
	// outstanding tasks have been cleaned up
	done chan struct{}
}

func (s *eremeticScheduler) newTask(offer *mesos.Offer, spec *types.EremeticTask) *mesos.TaskInfo {
	return createTaskInfo(spec, offer)
}

// Registered is called when the Scheduler is Registered
func (s *eremeticScheduler) Registered(_ sched.SchedulerDriver, frameworkID *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Debugf("Framework %s registered with master %s", frameworkID.GetValue(), masterInfo.GetHostname())
}

// Reregistered is called when the Scheduler is Reregistered
func (s *eremeticScheduler) Reregistered(_ sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Debugf("Framework re-registered with master %s", masterInfo)
}

// Disconnected is called when the Scheduler is Disconnected
func (s *eremeticScheduler) Disconnected(sched.SchedulerDriver) {
	log.Debugf("Framework disconnected with master")
}

// ResourceOffers handles the Resource Offers
func (s *eremeticScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log.Tracef("Received %d resource offers", len(offers))
	for _, offer := range offers {
		select {
		case <-s.shutdown:
			log.Infof("Shutting down: declining offer on [%s]", offer.Hostname)
			driver.DeclineOffer(offer.Id, defaultFilter)
			continue
		case tid := <-s.tasks:
			log.Debugf("Preparing to launch task %s with offer %s", tid, offer.Id.GetValue())
			t, _ := database.ReadTask(tid)
			task := s.newTask(offer, &t)
			database.PutTask(&t)
			driver.LaunchTasks([]*mesos.OfferID{offer.Id}, []*mesos.TaskInfo{task}, defaultFilter)
			continue
		default:
		}

		log.Trace("No tasks to launch. Declining offer.")
		driver.DeclineOffer(offer.Id, defaultFilter)
	}
}

// StatusUpdate takes care of updating the status
func (s *eremeticScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Debugf("Received task status [%s] for task [%s]", status.State.String(), *status.TaskId.Value)

	updateStatusForTask(status)
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

	database.PutTask(&task)
	s.tasks <- task.ID
	return task.ID, nil
}

func updateStatusForTask(status *mesos.TaskStatus) {
	id := status.TaskId.GetValue()
	log.Debugf("TaskId [%s] status [%s]", id, status.State)
	task, _ := database.ReadTask(id)
	newStatus := types.Status{
		Status: status.State.String(),
		Time:   time.Now().Unix(),
	}
	task.Status = append([]types.Status{newStatus}, task.Status...)

	database.PutTask(&task)
}
