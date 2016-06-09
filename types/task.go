package types

import (
	"fmt"
	"time"

	"github.com/m4rw3r/uuid"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type Status struct {
	Time   int64  `json:"time"`
	Status string `json:"status"`
}

type EremeticTask struct {
	TaskCPUs          float64           `json:"task_cpus"`
	TaskMem           float64           `json:"task_mem"`
	Command           string            `json:"command"`
	User              string            `json:"user"`
	Environment       map[string]string `json:"env"`
	MaskedEnvironment map[string]string `json:"masked_env"`
	Image             string            `json:"image"`
	Volumes           []Volume          `json:"volumes"`
	Status            []Status          `json:"status"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	FrameworkId       string            `json:"framework_id"`
	SlaveId           string            `json:"slave_id"`
	Hostname          string            `json:"hostname"`
	Retry             int               `json:"retry"`
	CallbackURI       string            `json:"callback_uri"`
	URIs              []string          `json:"uris"`
	SandboxPath       string            `json:"sandbox_path"`
	AgentIP           string            `json:"agent_ip"`
	AgentPort         int32             `json:"agent_port"`
}

func NewEremeticTask(request Request, name string) (EremeticTask, error) {
	uuid, err := uuid.V4()
	if err != nil {
		return EremeticTask{}, err
	}

	taskID := fmt.Sprintf("eremetic-task.%s", uuid)

	status := []Status{
		Status{
			Status: mesos.TaskState_TASK_STAGING.String(),
			Time:   time.Now().Unix(),
		},
	}

	task := EremeticTask{
		ID:                taskID,
		TaskCPUs:          request.TaskCPUs,
		TaskMem:           request.TaskMem,
		Name:              name,
		Status:            status,
		Command:           request.Command,
		User:              "root",
		Environment:       request.Environment,
		MaskedEnvironment: request.MaskedEnvironment,
		Image:             request.DockerImage,
		Volumes:           request.Volumes,
		CallbackURI:       request.CallbackURI,
		URIs:              request.URIs,
	}
	return task, nil
}

func (task *EremeticTask) WasRunning() bool {
	for _, s := range task.Status {
		if s.Status == mesos.TaskState_TASK_RUNNING.String() {
			return true
		}
	}
	return false
}

func (task *EremeticTask) IsTerminated() bool {
	if len(task.Status) == 0 {
		return true
	}
	st := task.Status[len(task.Status)-1]
	return IsTerminalString(st.Status)
}

func (task *EremeticTask) IsRunning() bool {
	if len(task.Status) == 0 {
		return false
	}
	st := task.Status[len(task.Status)-1]
	return st.Status == mesos.TaskState_TASK_RUNNING.String()
}

func (task *EremeticTask) LastUpdated() time.Time {
	if len(task.Status) == 0 {
		return time.Unix(0, 0)
	}
	st := task.Status[len(task.Status)-1]
	return time.Unix(st.Time, 0)
}

func (task *EremeticTask) UpdateStatus(status Status) {
	task.Status = append(task.Status, status)
}
