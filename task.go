package eremetic

import (
	"fmt"
	"strings"
	"time"

	"github.com/pborman/uuid"
)

// TaskState defines the valid task states.
type TaskState string

// Valid task states
const (
	// Standard mesos states
	TaskStaging  TaskState = "TASK_STAGING"
	TaskStarting TaskState = "TASK_STARTING"
	TaskRunning  TaskState = "TASK_RUNNING"
	TaskFinished TaskState = "TASK_FINISHED"
	TaskFailed   TaskState = "TASK_FAILED"
	TaskKilled   TaskState = "TASK_KILLED"
	TaskLost     TaskState = "TASK_LOST"
	TaskError    TaskState = "TASK_ERROR"

	// Custom eremetic states
	TaskQueued      TaskState = "TASK_QUEUED"
	TaskTerminating TaskState = "TASK_TERMINATING"
)

// IsTerminal takes a string representation of a state and returns whether it
// is terminal or not.
func IsTerminal(state TaskState) bool {
	switch state {
	case "TASK_LOST", "TASK_KILLED", "TASK_FAILED", "TASK_FINISHED":
		return true
	default:
		return false
	}
}

func (s TaskState) String() string {
	return string(s)
}

// Status represents the task status at a given time.
type Status struct {
	Time   int64     `json:"time"`
	Status TaskState `json:"status"`
}

// Volume is a mapping between ContainerPath and HostPath, to allow Docker
// to mount volumes.
type Volume struct {
	ContainerPath string `json:"container_path"`
	HostPath      string `json:"host_path"`
}

// Port defines a port mapping.
type Port struct {
	ContainerPort uint32 `json:"container_port"`
	HostPort      uint32 `json:"host_port"`
	Protocol      string `json:"protocol"`
}

// SlaveConstraint is a constraint that is validated for each slave when
// determining where to schedule a task.
type SlaveConstraint struct {
	AttributeName  string `json:"attribute_name"`
	AttributeValue string `json:"attribute_value"`
}

// URI holds meta-data for a sandbox resource.
type URI struct {
	URI        string `json:"uri"`
	Executable bool   `json:"executable"`
	Extract    bool   `json:"extract"`
	Cache      bool   `json:"cache"`
}

// Task defines the properties of a scheduled task.
type Task struct {
	TaskCPUs          float64           `json:"task_cpus"`
	TaskMem           float64           `json:"task_mem"`
	Command           string            `json:"command"`
	Args              []string          `json:"args"`
	User              string            `json:"user"`
	Environment       map[string]string `json:"env"`
	MaskedEnvironment map[string]string `json:"masked_env"`
	Image             string            `json:"image"`
	Volumes           []Volume          `json:"volumes"`
	Ports             []Port            `json:"ports"`
	Status            []Status          `json:"status"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Network           string            `json:"network"`
	DNS               string            `json:"dns"`
	FrameworkID       string            `json:"framework_id"`
	SlaveID           string            `json:"slave_id"`
	SlaveConstraints  []SlaveConstraint `json:"slave_constraints"`
	Hostname          string            `json:"hostname"`
	Retry             int               `json:"retry"`
	CallbackURI       string            `json:"callback_uri"`
	SandboxPath       string            `json:"sandbox_path"`
	AgentIP           string            `json:"agent_ip"`
	AgentPort         int32             `json:"agent_port"`
	ForcePullImage    bool              `json:"force_pull_image"`
	FetchURIs         []URI             `json:"fetch"`
}

func isArchive(url string) bool {
	var archiveSfx = []string{".tgz", ".tar.gz", ".tbz2", ".tar.bz2", ".txz", ".tar.xz", ".zip"}
	for _, s := range archiveSfx {
		if strings.HasSuffix(url, s) {
			return true
		}
	}
	return false
}

func mergeURIs(request Request) []URI {
	var URIs []URI
	for _, v := range request.URIs {
		URIs = append(URIs, URI{
			URI:        v,
			Extract:    isArchive(v),
			Cache:      false,
			Executable: false,
		})
	}
	for _, v := range request.Fetch {
		URIs = append(URIs, URI{
			URI:        v.URI,
			Extract:    v.Extract,
			Cache:      v.Cache,
			Executable: v.Executable,
		})
	}
	return URIs
}

// Request represents the structure of a job request
type Request struct {
	TaskCPUs          float64           `json:"task_cpus"`
	TaskMem           float64           `json:"task_mem"`
	DockerImage       string            `json:"docker_image"`
	Command           string            `json:"command"`
	Args              []string          `json:"args"`
	Volumes           []Volume          `json:"volumes"`
	Ports             []Port            `json:"ports"`
	Network           string            `json:"network"`
	DNS               string            `json:"dns"`
	Environment       map[string]string `json:"env"`
	MaskedEnvironment map[string]string `json:"masked_env"`
	SlaveConstraints  []SlaveConstraint `json:"slave_constraints"`
	CallbackURI       string            `json:"callback_uri"`
	URIs              []string          `json:"uris"`
	Fetch             []URI             `json:"fetch"`
	ForcePullImage    bool              `json:"force_pull_image"`
}

// NewTask returns a new instance of a Task.
func NewTask(request Request, name string) (Task, error) {
	taskID := fmt.Sprintf("eremetic-task.%s", uuid.New())

	status := []Status{
		Status{
			Status: TaskQueued,
			Time:   time.Now().Unix(),
		},
	}

	task := Task{
		ID:                taskID,
		TaskCPUs:          request.TaskCPUs,
		TaskMem:           request.TaskMem,
		Name:              name,
		Network:           request.Network,
		DNS:               request.DNS,
		Status:            status,
		Command:           request.Command,
		Args:              request.Args,
		User:              "root",
		Environment:       request.Environment,
		MaskedEnvironment: request.MaskedEnvironment,
		SlaveConstraints:  request.SlaveConstraints,
		Image:             request.DockerImage,
		Volumes:           request.Volumes,
		Ports:             request.Ports,
		CallbackURI:       request.CallbackURI,
		ForcePullImage:    request.ForcePullImage,
		FetchURIs:         mergeURIs(request),
	}
	return task, nil
}

// WasRunning returns whether the task was running at some point.
func (task *Task) WasRunning() bool {
	for _, s := range task.Status {
		if s.Status == TaskRunning {
			return true
		}
	}
	return false
}

// IsTerminated returns whether the task has been terminated.
func (task *Task) IsTerminated() bool {
	st := task.CurrentStatus()
	if st == "" {
		return true
	}
	return IsTerminal(st)
}

func (task *Task) IsWaiting() bool {
	return task.CurrentStatus() == TaskQueued
}

func (task *Task) IsTerminating() bool {
	return task.CurrentStatus() == TaskTerminating
}

// IsRunning returns whether the task is currently running.
func (task *Task) IsRunning() bool {
	return task.CurrentStatus() == TaskRunning
}

// CurrentStatus returns the current TaskState
func (task *Task) CurrentStatus() TaskState {
	if len(task.Status) == 0 {
		return ""
	}
	return task.Status[len(task.Status)-1].Status
}

// LastUpdated returns the time of the latest status update.
func (task *Task) LastUpdated() time.Time {
	if len(task.Status) == 0 {
		return time.Unix(0, 0)
	}
	st := task.Status[len(task.Status)-1]
	return time.Unix(st.Time, 0)
}

// UpdateStatus updates the current task status.
func (task *Task) UpdateStatus(status Status) {
	task.Status = append(task.Status, status)
}
