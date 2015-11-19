package types

import (
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
}

func (task *EremeticTask) WasRunning() bool {
	for _, s := range task.Status {
		if s.Status == mesos.TaskState_TASK_RUNNING.String() {
			return true
		}
	}
	return false
}
