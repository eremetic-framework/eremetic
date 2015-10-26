package types

type volume struct {
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
	Volumes     []volume          `json:"volumes"`
	Environment map[string]string `json:"env"`
}
