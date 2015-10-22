package handler

import (
	"fmt"
	log "github.com/dmuth/google-go-log4go"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	"github.com/m4rw3r/uuid"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

type eremeticTask struct {
	taskCPUs  float64
	taskMem   float64
	command   *mesos.CommandInfo
	container *mesos.ContainerInfo
}

func createEremeticTask(request types.Request) eremeticTask {
	task := eremeticTask{
		taskCPUs: request.TaskCPUs,
		taskMem:  request.TaskMem,
		command: &mesos.CommandInfo{
			Value: proto.String(request.Command),
			User:  proto.String("root"),
		},
		container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: proto.String(request.DockerImage),
			},
		},
	}
	return task
}

func createTaskInfo(task *eremeticTask, taskID int, offer *mesos.Offer) *mesos.TaskInfo {
	randId, err := uuid.V4()
	if err != nil {
		log.Error("Could not create random Id")
		os.Exit(1)
	}
	id := fmt.Sprintf("eremetic-task.%s", randId.String())

	return &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(id),
		},
		SlaveId:   offer.SlaveId,
		Name:      proto.String("Eremetic task " + string(taskID)),
		Command:   task.command,
		Container: task.container,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.taskCPUs),
			mesosutil.NewScalarResource("mem", task.taskMem),
		},
	}
}
