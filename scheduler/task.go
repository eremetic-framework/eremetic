package scheduler

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/klarna/eremetic/types"
	"github.com/m4rw3r/uuid"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

func createID(taskID string) string {
	return fmt.Sprintf("eremetic-task.%s", taskID)
}

func createEremeticTask(request types.Request) (types.EremeticTask, error) {
	randId, err := uuid.V4()
	if err != nil {
		return types.EremeticTask{}, err
	}
	taskId := createID(randId.String())

	status := []types.Status{
		types.Status{
			Status: mesos.TaskState_TASK_STAGING.String(),
			Time:   time.Now().Unix(),
		},
	}

	task := types.EremeticTask{
		ID:          taskId,
		TaskCPUs:    request.TaskCPUs,
		TaskMem:     request.TaskMem,
		Name:        request.Name,
		Status:      status,
		Command:     request.Command,
		User:        "root",
		Environment: request.Environment,
		Image:       request.DockerImage,
		Volumes:     request.Volumes,
		CallbackURI: request.CallbackURI,
	}
	return task, nil
}

func createTaskInfo(task types.EremeticTask, offer *mesos.Offer) (types.EremeticTask, *mesos.TaskInfo) {
	task.FrameworkId = *offer.FrameworkId.Value
	task.SlaveId = *offer.SlaveId.Value
	task.Hostname = *offer.Hostname

	var environment []*mesos.Environment_Variable
	for k, v := range task.Environment {
		environment = append(environment, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	environment = append(environment, &mesos.Environment_Variable{
		Name:  proto.String("MESOS_TASK_ID"),
		Value: proto.String(task.ID),
	})

	var volumes []*mesos.Volume
	for _, v := range task.Volumes {
		volumes = append(volumes, &mesos.Volume{
			Mode:          mesos.Volume_RW.Enum(),
			ContainerPath: proto.String(v.ContainerPath),
			HostPath:      proto.String(v.HostPath),
		})
	}

	return task, &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		SlaveId: offer.SlaveId,
		Name:    proto.String(task.Name),
		Command: &mesos.CommandInfo{
			Value: proto.String(task.Command),
			User:  proto.String(task.User),
			Environment: &mesos.Environment{
				Variables: environment,
			},
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: proto.String(task.Image),
			},
			Volumes: volumes,
		},
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.TaskCPUs),
			mesosutil.NewScalarResource("mem", task.TaskMem),
		},
	}
}
