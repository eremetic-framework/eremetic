package eremetic

import (
	"fmt"
	"strings"
	"time"

	"github.com/pborman/uuid"
)

type TaskState string

const (
	// Standard mesos states
	TaskState_TASK_STAGING  TaskState = "TASK_STAGING"
	TaskState_TASK_STARTING TaskState = "TASK_STARTING"
	TaskState_TASK_RUNNING  TaskState = "TASK_RUNNING"
	TaskState_TASK_FINISHED TaskState = "TASK_FINISHED"
	TaskState_TASK_FAILED   TaskState = "TASK_FAILED"
	TaskState_TASK_KILLED   TaskState = "TASK_KILLED"
	TaskState_TASK_LOST     TaskState = "TASK_LOST"
	TaskState_TASK_ERROR    TaskState = "TASK_ERROR"
	// Custom eremetic states
	TaskState_TASK_QUEUED TaskState = "TASK_QUEUED"
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

type Port struct {
	ContainerPort uint32 `json:"container_port"`
	HostPort      uint32 `json:"host_port"`
	Protocol      string `json:"protocol"`
}

type SlaveConstraint struct {
	AttributeName  string `json:"attribute_name"`
	AttributeValue string `json:"attribute_value"`
}

type URI struct {
	URI        string `json:"uri"`
	Executable bool   `json:"executable"`
	Extract    bool   `json:"extract"`
	Cache      bool   `json:"cache"`
}

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
	FrameworkId       string            `json:"framework_id"`
	SlaveId           string            `json:"slave_id"`
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
	Environment       map[string]string `json:"env"`
	MaskedEnvironment map[string]string `json:"masked_env"`
	SlaveConstraints  []SlaveConstraint `json:"slave_constraints"`
	CallbackURI       string            `json:"callback_uri"`
	URIs              []string          `json:"uris"`
	Fetch             []URI             `json:"fetch"`
	ForcePullImage    bool              `json:"force_pull_image"`
}

func NewTask(request Request, name string) (Task, error) {
	taskID := fmt.Sprintf("eremetic-task.%s", uuid.New())

	status := []Status{
		Status{
			Status: TaskState_TASK_QUEUED,
			Time:   time.Now().Unix(),
		},
	}

	task := Task{
		ID:                taskID,
		TaskCPUs:          request.TaskCPUs,
		TaskMem:           request.TaskMem,
		Name:              name,
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

func (task *Task) WasRunning() bool {
	for _, s := range task.Status {
		if s.Status == TaskState_TASK_RUNNING {
			return true
		}
	}
	return false
}

func (task *Task) IsTerminated() bool {
	if len(task.Status) == 0 {
		return true
	}
	st := task.Status[len(task.Status)-1]
	return IsTerminal(st.Status)
}

func (task *Task) IsRunning() bool {
	if len(task.Status) == 0 {
		return false
	}
	st := task.Status[len(task.Status)-1]
	return st.Status == TaskState_TASK_RUNNING
}

func (task *Task) LastUpdated() time.Time {
	if len(task.Status) == 0 {
		return time.Unix(0, 0)
	}
	st := task.Status[len(task.Status)-1]
	return time.Unix(st.Time, 0)
}

func (task *Task) UpdateStatus(status Status) {
	task.Status = append(task.Status, status)
}
