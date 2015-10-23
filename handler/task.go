package handler

import (
	"time"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

type eremeticTask struct {
	TaskCPUs  float64              `json:"task_cpus"`
	TaskMem   float64              `json:"task_mem"`
	Command   *mesos.CommandInfo   `json:"command"`
	Container *mesos.ContainerInfo `json:"container"`
	Status    string               `json:"status"`
	ID        string               `json:"-"`
	deleteAt  time.Time
}

var runningTasks map[string]eremeticTask

func createEremeticTask(request types.Request) eremeticTask {
	task := eremeticTask{
		TaskCPUs: request.TaskCPUs,
		TaskMem:  request.TaskMem,
		ID:       request.TaskID,
		Command: &mesos.CommandInfo{
			Value: proto.String(request.Command),
			User:  proto.String("root"),
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: proto.String(request.DockerImage),
			},
		},
	}
	return task
}

func createTaskInfo(task *eremeticTask, taskID int, offer *mesos.Offer) *mesos.TaskInfo {

	return &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		SlaveId:   offer.SlaveId,
		Name:      proto.String("Eremetic task " + string(taskID)),
		Command:   task.Command,
		Container: task.Container,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.TaskCPUs),
			mesosutil.NewScalarResource("mem", task.TaskMem),
		},
	}
}
