package api

import (
	"github.com/eremetic-framework/eremetic"
)

// TaskV0 defines the deprecated json-structure for the properties of a scheduled task.
type TaskV0 struct {
	TaskCPUs          float64                    `json:"task_cpus"`
	TaskMem           float64                    `json:"task_mem"`
	Command           string                     `json:"command"`
	Args              []string                   `json:"args"`
	User              string                     `json:"user"`
	Environment       map[string]string          `json:"env"`
	MaskedEnvironment map[string]string          `json:"masked_env"`
	Labels            map[string]string          `json:"labels"`
	Image             string                     `json:"image"`
	Volumes           []eremetic.Volume          `json:"volumes"`
	Ports             []eremetic.Port            `json:"ports"`
	Status            []eremetic.Status          `json:"status"`
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	Network           string                     `json:"network"`
	DNS               string                     `json:"dns"`
	FrameworkID       string                     `json:"framework_id"`
	AgentID           string                     `json:"slave_id"`
	AgentConstraints  []eremetic.AgentConstraint `json:"slave_constraints"`
	Hostname          string                     `json:"hostname"`
	Retry             int                        `json:"retry"`
	CallbackURI       string                     `json:"callback_uri"`
	SandboxPath       string                     `json:"sandbox_path"`
	AgentIP           string                     `json:"agent_ip"`
	AgentPort         int32                      `json:"agent_port"`
	ForcePullImage    bool                       `json:"force_pull_image"`
	Privileged        bool                       `json:"privileged"`
	FetchURIs         []eremetic.URI             `json:"fetch"`
}

// TaskV0FromTask is needed for Go versions < 1.8
// In go 1.8, TaskV0(Task) would work instead
func TaskV0FromTask(task *eremetic.Task) TaskV0 {
	return TaskV0{
		TaskCPUs:          task.TaskCPUs,
		TaskMem:           task.TaskMem,
		Command:           task.Command,
		Args:              task.Args,
		User:              task.User,
		Environment:       task.Environment,
		MaskedEnvironment: task.MaskedEnvironment,
		Labels:            task.Labels,
		Image:             task.Image,
		Volumes:           task.Volumes,
		Ports:             task.Ports,
		Status:            task.Status,
		ID:                task.ID,
		Name:              task.Name,
		Network:           task.Network,
		DNS:               task.DNS,
		FrameworkID:       task.FrameworkID,
		AgentID:           task.AgentID,
		AgentConstraints:  task.AgentConstraints,
		Hostname:          task.Hostname,
		Retry:             task.Retry,
		CallbackURI:       task.CallbackURI,
		SandboxPath:       task.SandboxPath,
		AgentIP:           task.AgentIP,
		AgentPort:         task.AgentPort,
		ForcePullImage:    task.ForcePullImage,
		Privileged:        task.Privileged,
		FetchURIs:         task.FetchURIs,
	}
}

// TaskFromV0 is needed for Go versions < 1.8
// In go 1.8, Task(TaskV0) would work instead
func TaskFromV0(task *TaskV0) eremetic.Task {
	return eremetic.Task{
		TaskCPUs:          task.TaskCPUs,
		TaskMem:           task.TaskMem,
		Command:           task.Command,
		Args:              task.Args,
		User:              task.User,
		Environment:       task.Environment,
		MaskedEnvironment: task.MaskedEnvironment,
		Labels:             task.Labels,
		Image:             task.Image,
		Volumes:           task.Volumes,
		Ports:             task.Ports,
		Status:            task.Status,
		ID:                task.ID,
		Name:              task.Name,
		Network:           task.Network,
		DNS:               task.DNS,
		FrameworkID:       task.FrameworkID,
		AgentID:           task.AgentID,
		AgentConstraints:  task.AgentConstraints,
		Hostname:          task.Hostname,
		Retry:             task.Retry,
		CallbackURI:       task.CallbackURI,
		SandboxPath:       task.SandboxPath,
		AgentIP:           task.AgentIP,
		AgentPort:         task.AgentPort,
		ForcePullImage:    task.ForcePullImage,
		Privileged:        task.Privileged,
		FetchURIs:         task.FetchURIs,
	}
}

// RequestV0 represents the old deprecated json-structure of a job request
type RequestV0 struct {
	TaskCPUs          float64                    `json:"task_cpus"`
	TaskMem           float64                    `json:"task_mem"`
	DockerImage       string                     `json:"docker_image"`
	Command           string                     `json:"command"`
	Args              []string                   `json:"args"`
	Volumes           []eremetic.Volume          `json:"volumes"`
	Ports             []eremetic.Port            `json:"ports"`
	Name              string                     `json:"name"`
	Network           string                     `json:"network"`
	DNS               string                     `json:"dns"`
	Environment       map[string]string          `json:"env"`
	MaskedEnvironment map[string]string          `json:"masked_env"`
	Labels            map[string]string          `json:"labels"`
	AgentConstraints  []eremetic.AgentConstraint `json:"slave_constraints"`
	CallbackURI       string                     `json:"callback_uri"`
	URIs              []string                   `json:"uris"`
	Fetch             []eremetic.URI             `json:"fetch"`
	ForcePullImage    bool                       `json:"force_pull_image"`
	Privileged        bool                       `json:"privileged"`
}

// RequestFromV0 is needed for Go versions < 1.8
// In go 1.8, Request(RequestV0) would work instead
func RequestFromV0(req RequestV0) eremetic.Request {
	return eremetic.Request{
		TaskCPUs:          req.TaskCPUs,
		TaskMem:           req.TaskMem,
		DockerImage:       req.DockerImage,
		Command:           req.Command,
		Args:              req.Args,
		Volumes:           req.Volumes,
		Ports:             req.Ports,
		Name:              req.Name,
		Network:           req.Network,
		DNS:               req.DNS,
		Environment:       req.Environment,
		MaskedEnvironment: req.MaskedEnvironment,
		Labels:            req.Labels,
		AgentConstraints:  req.AgentConstraints,
		CallbackURI:       req.CallbackURI,
		URIs:              req.URIs,
		Fetch:             req.Fetch,
		ForcePullImage:    req.ForcePullImage,
		Privileged:        req.Privileged,
	}
}
