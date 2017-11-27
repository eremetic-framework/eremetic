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

// IsActive takes a string representation of a state and returns whether it
// is active or not.
func IsActive(state TaskState) bool {
	switch state {
	case "TASK_STAGING", "TASK_STARTING", "TASK_RUNNING", "TASK_ERROR", "TASK_TERMINATING":
		return true
	default:
		return false
	}
}

// IsEnqueued takes a string representation of a state and returns whether it
// is enqueued or not.
func IsEnqueued(state TaskState) bool {
	switch state {
	case "TASK_QUEUED":
		return true
	default:
		return false
	}
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

// AgentConstraint is a constraint that is validated for each agent when
// determining where to schedule a task.
type AgentConstraint struct {
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

// Task represents the internal structure of a Task object
type Task struct {
	TaskCPUs          float64
	TaskMem           float64
	Command           string
	Args              []string
	User              string
	Environment       map[string]string
	MaskedEnvironment map[string]string
	Labels            map[string]string
	Image             string
	Volumes           []Volume
	VolumesFrom       []string
	Ports             []Port
	Status            []Status
	ID                string
	Name              string
	Network           string
	DNS               string
	FrameworkID       string
	AgentID           string
	AgentConstraints  []AgentConstraint
	Hostname          string
	Retry             int
	CallbackURI       string
	SandboxPath       string
	AgentIP           string
	AgentPort         int32
	ForcePullImage    bool
	Privileged        bool
	FetchURIs         []URI
}

// TaskFilter represents the query param state
type TaskFilter struct {
	Name  string `schema:"name"`
	State string `schema:"state"`
}

// Possible states for the TaskFilter. And the default state
const (
	DefaultTaskFilterState = "active,queued"
	TerminatedState        = "terminated"
	ActiveState            = "active"
	QueuedState            = "queued"
)

// IsArchive is used to determine whether a url is an archive or not
func IsArchive(url string) bool {
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
			Extract:    IsArchive(v),
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

// Request is the internal structure of a Request
type Request struct {
	TaskCPUs          float64
	TaskMem           float64
	DockerImage       string
	Command           string
	Args              []string
	Volumes           []Volume
	VolumesFrom       []string
	Ports             []Port
	Name              string
	Network           string
	DNS               string
	Environment       map[string]string
	MaskedEnvironment map[string]string
	Labels            map[string]string
	AgentConstraints  []AgentConstraint
	CallbackURI       string
	URIs              []string
	Fetch             []URI
	ForcePullImage    bool
	Privileged        bool
}

// NewTask returns a new instance of a Task.
func NewTask(request Request) (Task, error) {
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
		Name:              request.Name,
		Network:           request.Network,
		DNS:               request.DNS,
		Status:            status,
		Command:           request.Command,
		Args:              request.Args,
		User:              "root",
		Environment:       request.Environment,
		MaskedEnvironment: request.MaskedEnvironment,
		AgentConstraints:  request.AgentConstraints,
		Labels:            request.Labels,
		Image:             request.DockerImage,
		Volumes:           request.Volumes,
		VolumesFrom:       request.VolumesFrom,
		Ports:             request.Ports,
		CallbackURI:       request.CallbackURI,
		ForcePullImage:    request.ForcePullImage,
		Privileged:        request.Privileged,
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

// IsActive returns whether the task is still active.
func (task *Task) IsActive() bool {
	return IsActive(task.CurrentStatus())
}

// IsEnqueued returns whether the task is in queue.
func (task *Task) IsEnqueued() bool {
	return IsEnqueued(task.CurrentStatus())
}

// IsTerminating returns whether a task is in the process of terminating
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

// Match the conditions of TaskFilter with the current task
func (filter TaskFilter) Match(task *Task) bool {
	if len(filter.Name) > 0 {
		if filter.Name != task.Name {
			return false
		}
	}
	if len(filter.State) > 0 {
		if !taskHasAnyState(task, filter.State) {
			return false
		}
	}
	return true
}
func taskHasAnyState(task *Task, states string) bool {
	result := false
	for _, state := range strings.Split(states, ",") {
		switch state {
		case "active":
			result = result || task.IsActive()
		case "terminated":
			result = result || task.IsTerminated()
		case "queued":
			result = result || task.IsEnqueued()
		default:
			result = result || false
		}
	}
	return result
}
