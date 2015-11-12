package handler

import (
	"fmt"
	"time"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	"github.com/m4rw3r/uuid"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

type eremeticTask struct {
	TaskCPUs    float64              `json:"task_cpus"`
	TaskMem     float64              `json:"task_mem"`
	Command     *mesos.CommandInfo   `json:"command"`
	Container   *mesos.ContainerInfo `json:"container"`
	Status      string               `json:"status"`
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	FrameworkId string               `json:"framework_id"`
	SlaveId     string               `json:"slave_id"`
	Hostname    string               `json:"hostname"`
	deleteAt    time.Time
}

var runningTasks map[string]eremeticTask

func createID(taskID string) string {
	return fmt.Sprintf("eremetic-task.%s", taskID)
}

func createEremeticTask(request types.Request) (eremeticTask, error) {
	randId, err := uuid.V4()
	if err != nil {
		return eremeticTask{}, err
	}
	taskId := createID(randId.String())

	var volumes []*mesos.Volume
	for _, v := range request.Volumes {
		volumes = append(volumes, &mesos.Volume{
			Mode:          mesos.Volume_RW.Enum(),
			ContainerPath: proto.String(v.ContainerPath),
			HostPath:      proto.String(v.HostPath),
		})
	}

	var environment []*mesos.Environment_Variable
	for k, v := range request.Environment {
		environment = append(environment, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	environment = append(environment, &mesos.Environment_Variable{
		Name:  proto.String("MESOS_TASK_ID"),
		Value: proto.String(taskId),
	})

	task := eremeticTask{
		ID:       taskId,
		TaskCPUs: request.TaskCPUs,
		TaskMem:  request.TaskMem,
		Name:     request.Name,
		Status:   mesos.TaskState_TASK_STAGING.String(),
		Command: &mesos.CommandInfo{
			Value: proto.String(request.Command),
			User:  proto.String("root"),
			Environment: &mesos.Environment{
				Variables: environment,
			},
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image: proto.String(request.DockerImage),
			},
			Volumes: volumes,
		},
	}
	return task, nil
}

func createTaskInfo(task *eremeticTask, offer *mesos.Offer) *mesos.TaskInfo {
	task.FrameworkId = *offer.FrameworkId.Value
	task.SlaveId = *offer.SlaveId.Value
	task.Hostname = *offer.Hostname

	return &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		SlaveId:   offer.SlaveId,
		Name:      proto.String(task.Name),
		Command:   task.Command,
		Container: task.Container,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.TaskCPUs),
			mesosutil.NewScalarResource("mem", task.TaskMem),
		},
	}
}
