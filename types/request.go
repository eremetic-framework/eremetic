package types

// Volume is a mapping between ContainerPath and HostPath, to allow Docker
// to mount volumes.
type Volume struct {
	ContainerPath string `json:"container_path"`
	HostPath      string `json:"host_path"`
}

// Request represents the structure of a job request
type Request struct {
	TaskCPUs    float64           `json:"task_cpus"`
	TaskMem     float64           `json:"task_mem"`
	DockerImage string            `json:"docker_image"`
	Command     string            `json:"command"`
	TaskID      string            `json:"-"`
	Volumes     []Volume          `json:"volumes"`
	Environment map[string]string `json:"env"`
}
