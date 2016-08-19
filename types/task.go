package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/m4rw3r/uuid"
)

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

func NewEremeticTask(request Request, name string) (EremeticTask, error) {
	uuid, err := uuid.V4()
	if err != nil {
		return EremeticTask{}, err
	}

	taskID := fmt.Sprintf("eremetic-task.%s", uuid)

	status := []Status{
		Status{
			Status: TaskState_TASK_QUEUED,
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
		SlaveConstraints:  request.SlaveConstraints,
		Image:             request.DockerImage,
		Volumes:           request.Volumes,
		CallbackURI:       request.CallbackURI,
		ForcePullImage:    request.ForcePullImage,
		FetchURIs:         mergeURIs(request),
	}
	return task, nil
}

func (task *EremeticTask) WasRunning() bool {
	for _, s := range task.Status {
		if s.Status == TaskState_TASK_RUNNING {
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
	return IsTerminal(st.Status)
}

func (task *EremeticTask) IsRunning() bool {
	if len(task.Status) == 0 {
		return false
	}
	st := task.Status[len(task.Status)-1]
	return st.Status == TaskState_TASK_RUNNING
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

func (task *EremeticTask) IsExpired(retentionPeriod int64) bool {
	if task.IsTerminated() {
		st := task.Status[len(task.Status)-1]
		delta := time.Now().Unix() - st.Time
		if delta > retentionPeriod {
			return true
		}
	}
	return false
}
