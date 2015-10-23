package types

// Request represents the structure of a job request
type Request struct {
	TaskCPUs      float64 `json:"task_cpus"`
	TaskMem       float64 `json:"task_mem"`
	DockerImage   string  `json:"docker_image"`
	Command       string  `json:"command"`
	TasksToLaunch int     `json:"tasks_to_launch"`
	TaskID        string  `json:"-"`
}
