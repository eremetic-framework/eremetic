package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
)

var (
	defaultFilter = &mesos.Filters{RefuseSeconds: proto.Float64(1)}
)

// eremeticScheduler holds the structure of the Eremetic Scheduler
type eremeticScheduler struct {
	taskCPUs      float64
	taskMem       float64
	dockerImage   string
	command       string
	tasksToLaunch int
	tasksCreated  int
	tasksRunning  int

	eremeticExecutor *mesos.ExecutorInfo

	// This channel is closed when the program receives an interrupt,
	// signalling that the program should shut down.
	shutdown chan struct{}
	// This channel is closed after shutdown is closed, and only when all
	// outstanding tasks have been cleaned up
	done chan struct{}
}

func (s *eremeticScheduler) newTaskPrototype(offer *mesos.Offer) *mesos.TaskInfo {
	taskID := s.tasksCreated
	s.tasksCreated++
	return &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(fmt.Sprintf("Eremetic-%d: Running '%s' on '%s'", taskID, s.command, s.dockerImage)),
		},
		SlaveId: offer.SlaveId,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", s.taskCPUs),
			mesosutil.NewScalarResource("mem", s.taskMem),
		},
	}
}

func (s *eremeticScheduler) newTask(offer *mesos.Offer) *mesos.TaskInfo {
	task := s.newTaskPrototype(offer)
	task.Name = proto.String("EREMETIC_" + *task.TaskId.Value)
	task.Executor = s.eremeticExecutor
	return task
}

// Registered is called when the Scheduler is Registered
func (s *eremeticScheduler) Registered(
	_ sched.SchedulerDriver,
	frameworkID *mesos.FrameworkID,
	masterInfo *mesos.MasterInfo) {
	log.Printf("Framework %s registered with master %s", frameworkID, masterInfo)
}

// Reregistered is called when the Scheduler is Reregistered
func (s *eremeticScheduler) Reregistered(_ sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Printf("Framework re-registered with master %s", masterInfo)
}

// Disconnected is called when the Scheduler is Disconnected
func (s *eremeticScheduler) Disconnected(sched.SchedulerDriver) {
	log.Println("Framework disconnected with master")
}

// ResourceOffers handles the Resource Offers
func (s *eremeticScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log.Printf("Received %d resource offers", len(offers))
	for _, offer := range offers {
		select {
		case <-s.shutdown:
			log.Println("Shutting down: declining offer on [", offer.Hostname, "]")
			driver.DeclineOffer(offer.Id, defaultFilter)
			if s.tasksRunning == 0 {
				close(s.done)
			}
			continue
		default:
		}

		tasks := []*mesos.TaskInfo{}
		for s.tasksToLaunch > 0 {
			task := s.newTask(offer)
			tasks = append(tasks, task)
			s.tasksToLaunch--
		}

		if len(tasks) == 0 {
			log.Print("No tasks to launch. Declining offer.")
			driver.DeclineOffer(offer.Id, defaultFilter)
		} else {
			log.Printf("Launching %d tasks.", len(tasks))
			driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, defaultFilter)
		}
	}
}

// StatusUpdate takes care of updating the status
func (s *eremeticScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Printf("Received task status [%s] for task [%s]", types.NameFor(status.State), *status.TaskId.Value)

	if *status.State == mesos.TaskState_TASK_RUNNING {
		s.tasksRunning++
	} else if types.IsTerminal(status.State) {
		s.tasksRunning--
		if s.tasksRunning == 0 {
			select {
			case <-s.shutdown:
				close(s.done)
			default:
			}
		}
	}
}

func (s *eremeticScheduler) FrameworkMessage(
	driver sched.SchedulerDriver,
	executorID *mesos.ExecutorID,
	slaveID *mesos.SlaveID,
	message string) {

	log.Println("Getting a framework message")
	switch *executorID.Value {
	case *s.eremeticExecutor.ExecutorId.Value:
		log.Printf("Received framework message from renderer")
		var result interface{}
		err := json.Unmarshal([]byte(message), &result)
		if err != nil {
			log.Printf("Error deserializing Result: [%s]", err)
			return
		}
		log.Printf(
			"Appending [%s] to render results",
			result,
		)

	default:
		log.Printf("Received a framework message from some unknown source: %s", *executorID.Value)
	}
}

func (s *eremeticScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	log.Printf("Offer %s rescinded", offerID)
}
func (s *eremeticScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	log.Printf("Slave %s lost", slaveID)
}
func (s *eremeticScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	log.Printf("Executor %s on slave %s was lost", executorID, slaveID)
}

func (s *eremeticScheduler) Error(_ sched.SchedulerDriver, err string) {
	log.Printf("Receiving an error: %s", err)
}

// CreateeremeticScheduler creates a new scheduler for Rendler.
func createEremeticScheduler(request types.Request) *eremeticScheduler {
	s := &eremeticScheduler{
		taskCPUs:      request.TaskCPUs,
		taskMem:       request.TaskMem,
		dockerImage:   request.DockerImage,
		command:       request.Command,
		tasksToLaunch: request.TasksToLaunch,
		shutdown:      make(chan struct{}),
		done:          make(chan struct{}),
		eremeticExecutor: &mesos.ExecutorInfo{
			ExecutorId: &mesos.ExecutorID{Value: proto.String("eremetic-executor")},
			Command: &mesos.CommandInfo{
				Value: proto.String(request.Command),
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_DOCKER.Enum(),
				Docker: &mesos.ContainerInfo_DockerInfo{
					Image: proto.String(request.DockerImage),
				},
			},
			Name: proto.String("Eremetic"),
		},
	}
	return s
}
