package types

// Request represents the structure of a job request
type Request struct {
	TaskCPUs          float64           `json:"task_cpus"`
	TaskMem           float64           `json:"task_mem"`
	DockerImage       string            `json:"docker_image"`
	Command           string            `json:"command"`
	Name              string            `json:"-"`
	Volumes           []Volume          `json:"volumes"`
	Environment       map[string]string `json:"env"`
	MaskedEnvironment map[string]string `json:"masked_env"`
	SlaveConstraints  []SlaveConstraint `json:"slave_constraint"`
	CallbackURI       string            `json:"callback_uri"`
	URIs              []string          `json:"uris"`
}
