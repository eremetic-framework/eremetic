package api

import (
	"github.com/eremetic-framework/eremetic"
)

// TaskV1 defines the API V1 json-structure for the properties of a scheduled task.
type TaskV1 struct {
	TaskCPUs              float64                    `json:"cpu"`
	TaskMem               float64                    `json:"mem"`
	Command               string                     `json:"command"`
	Args                  []string                   `json:"args"`
	User                  string                     `json:"user"`
	Environment           map[string]string          `json:"env"`
	MaskedEnvironment     map[string]string          `json:"masked_env"`
	Image                 string                     `json:"image"`
	Volumes               []eremetic.Volume          `json:"volumes"`
	VolumesFromContainers []string                   `json:"volumes_from_containers"`
	Ports                 []eremetic.Port            `json:"ports"`
	Status                []eremetic.Status          `json:"status"`
	ID                    string                     `json:"id"`
	Name                  string                     `json:"name"`
	Network               string                     `json:"network"`
	DNS                   string                     `json:"dns"`
	FrameworkID           string                     `json:"framework_id"`
	AgentID               string                     `json:"agent_id"`
	AgentConstraints      []eremetic.AgentConstraint `json:"agent_constraints"`
	Hostname              string                     `json:"hostname"`
	Retry                 int                        `json:"retry"`
	CallbackURI           string                     `json:"callback_uri"`
	SandboxPath           string                     `json:"sandbox_path"`
	AgentIP               string                     `json:"agent_ip"`
	AgentPort             int32                      `json:"agent_port"`
	ForcePullImage        bool                       `json:"force_pull_image"`
	Privileged            bool                       `json:"privileged"`
	FetchURIs             []eremetic.URI             `json:"fetch"`
}

// TaskV1FromTask is needed for Go versions < 1.8
// In go 1.8, TaskV1(Task) would work instead
func TaskV1FromTask(task *eremetic.Task) TaskV1 {
	return TaskV1{
		TaskCPUs:              task.TaskCPUs,
		TaskMem:               task.TaskMem,
		Command:               task.Command,
		Args:                  task.Args,
		User:                  task.User,
		Environment:           task.Environment,
		MaskedEnvironment:     task.MaskedEnvironment,
		Image:                 task.Image,
		Volumes:               task.Volumes,
		VolumesFromContainers: task.VolumesFromContainers,
		Ports:            task.Ports,
		Status:           task.Status,
		ID:               task.ID,
		Name:             task.Name,
		Network:          task.Network,
		DNS:              task.DNS,
		FrameworkID:      task.FrameworkID,
		AgentID:          task.AgentID,
		AgentConstraints: task.AgentConstraints,
		Hostname:         task.Hostname,
		Retry:            task.Retry,
		CallbackURI:      task.CallbackURI,
		SandboxPath:      task.SandboxPath,
		AgentIP:          task.AgentIP,
		AgentPort:        task.AgentPort,
		ForcePullImage:   task.ForcePullImage,
		Privileged:       task.Privileged,
		FetchURIs:        task.FetchURIs,
	}
}

// TaskFromV1 is needed for Go versions < 1.8
// In go 1.8, Task(TaskV1) would work instead
func TaskFromV1(task *TaskV1) eremetic.Task {
	return eremetic.Task{
		TaskCPUs:              task.TaskCPUs,
		TaskMem:               task.TaskMem,
		Command:               task.Command,
		Args:                  task.Args,
		User:                  task.User,
		Environment:           task.Environment,
		MaskedEnvironment:     task.MaskedEnvironment,
		Image:                 task.Image,
		Volumes:               task.Volumes,
		VolumesFromContainers: task.VolumesFromContainers,
		Ports:            task.Ports,
		Status:           task.Status,
		ID:               task.ID,
		Name:             task.Name,
		Network:          task.Network,
		DNS:              task.DNS,
		FrameworkID:      task.FrameworkID,
		AgentID:          task.AgentID,
		AgentConstraints: task.AgentConstraints,
		Hostname:         task.Hostname,
		Retry:            task.Retry,
		CallbackURI:      task.CallbackURI,
		SandboxPath:      task.SandboxPath,
		AgentIP:          task.AgentIP,
		AgentPort:        task.AgentPort,
		ForcePullImage:   task.ForcePullImage,
		Privileged:       task.Privileged,
		FetchURIs:        task.FetchURIs,
	}
}

// RequestV1 represents the V1 json-structure of a job request
type RequestV1 struct {
	TaskCPUs              float64                    `json:"cpu"`
	TaskMem               float64                    `json:"mem"`
	DockerImage           string                     `json:"image"`
	Command               string                     `json:"command"`
	Args                  []string                   `json:"args"`
	Volumes               []eremetic.Volume          `json:"volumes"`
	VolumesFromContainers []string                   `json:"volumes_from_containers"`
	Ports                 []eremetic.Port            `json:"ports"`
	Network               string                     `json:"network"`
	DNS                   string                     `json:"dns"`
	Environment           map[string]string          `json:"env"`
	MaskedEnvironment     map[string]string          `json:"masked_env"`
	AgentConstraints      []eremetic.AgentConstraint `json:"agent_constraints"`
	CallbackURI           string                     `json:"callback_uri"`
	Fetch                 []eremetic.URI             `json:"fetch"`
	ForcePullImage        bool                       `json:"force_pull_image"`
	Privileged            bool                       `json:"privileged"`
}

// RequestFromV1 is needed for Go versions < 1.8
// In go 1.8, Request(RequestV1) would work instead
func RequestFromV1(req RequestV1) eremetic.Request {
	return eremetic.Request{
		TaskCPUs:              req.TaskCPUs,
		TaskMem:               req.TaskMem,
		DockerImage:           req.DockerImage,
		Command:               req.Command,
		Args:                  req.Args,
		Volumes:               req.Volumes,
		VolumesFromContainers: req.VolumesFromContainers,
		Ports:             req.Ports,
		Network:           req.Network,
		DNS:               req.DNS,
		Environment:       req.Environment,
		MaskedEnvironment: req.MaskedEnvironment,
		AgentConstraints:  req.AgentConstraints,
		CallbackURI:       req.CallbackURI,
		URIs:              []string{},
		Fetch:             req.Fetch,
		ForcePullImage:    req.ForcePullImage,
		Privileged:        req.Privileged,
	}
}
