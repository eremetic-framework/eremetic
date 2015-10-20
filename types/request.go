package types

// POST {
//  "task_cpus": 0.5,
//  "task_mem": 23.0,
//  "docker_image": 132465846eauoo6e54ae8ae3u,
//  "command": "hello",
// }

// Request represents the structure of a job request
type Request struct {
	TaskCPUs      float64 `json:"task_cpus"`
	TaskMem       float64 `json:"task_mem"`
	DockerImage   string  `json:"docker_image"`
	Command       string  `json:"command"`
	TasksToLaunch int     `json:"tasks_to_launch"`
}
