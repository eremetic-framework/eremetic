package handler

import (
	"time"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
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

func createEremeticTask(request types.Request) eremeticTask {
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

	task := eremeticTask{
		TaskCPUs: request.TaskCPUs,
		TaskMem:  request.TaskMem,
		ID:       request.TaskID,
		Name:     request.Name,
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
	return task
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
