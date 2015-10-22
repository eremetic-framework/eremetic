package handler

import (
	"fmt"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

type eremeticTask struct {
	taskCPUs    float64
	taskMem     float64
	dockerImage string
	command     string
	executor    *mesos.ExecutorInfo
}

func createEremeticTask(request types.Request) eremeticTask {
	task := eremeticTask{
		taskCPUs:    request.TaskCPUs,
		taskMem:     request.TaskMem,
		dockerImage: request.DockerImage,
		command:     request.Command,
		executor: &mesos.ExecutorInfo{
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
	return task
}

func createTaskInfo(task *eremeticTask, taskID int, offer *mesos.Offer) *mesos.TaskInfo {
	id := proto.String(fmt.Sprintf(
		"Eremetic-%d: Running '%s' on '%s'",
		taskID, task.command, task.dockerImage))

	return &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: id,
		},
		SlaveId:  offer.SlaveId,
		Name:     proto.String("EREMETIC_" + *id),
		Executor: task.executor,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.taskCPUs),
			mesosutil.NewScalarResource("mem", task.taskMem),
		},
	}
}
