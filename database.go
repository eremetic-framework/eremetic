package eremetic

import "encoding/json"

// TaskDB defines the functions needed by the database abstraction layer
type TaskDB interface {
	Clean() error
	Close()
	PutTask(task *Task) error
	ReadTask(id string) (Task, error)
	ReadUnmaskedTask(id string) (Task, error)
	ListNonTerminalTasks() ([]*Task, error)
}

const Masking = "*******"

func ApplyMask(task *Task) {
	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = Masking
	}
}

func Encode(task *Task) ([]byte, error) {
	encoded, err := json.Marshal(task)
	return []byte(encoded), err
}
