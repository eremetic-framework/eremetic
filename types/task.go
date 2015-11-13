package types

import (
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type EremeticTask struct {
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
}
