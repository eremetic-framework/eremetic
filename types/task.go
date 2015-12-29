package types

import (
	"time"

	mesos "github.com/mesos/mesos-go/mesosproto"
)

type Status struct {
	Time   int64  `json:"time"`
	Status string `json:"status"`
}

type EremeticTask struct {
	TaskCPUs    float64              `json:"task_cpus"`
	TaskMem     float64              `json:"task_mem"`
	Command     *mesos.CommandInfo   `json:"command"`
	Container   *mesos.ContainerInfo `json:"container"`
	Status      []Status             `json:"status"`
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	FrameworkId string               `json:"framework_id"`
	SlaveId     string               `json:"slave_id"`
	Hostname    string               `json:"hostname"`
	Retry       int                  `json:"retry"`
	CallbackURI string               `json:"callback_uri"`
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
