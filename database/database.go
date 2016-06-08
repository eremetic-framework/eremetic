package database

import "github.com/klarna/eremetic/types"

type TaskDB interface {
	Clean() error
	Close()
	PutTask(task *types.EremeticTask) error
	ReadTask(id string) (types.EremeticTask, error)
	ReadUnmaskedTask(id string) (types.EremeticTask, error)
	ListNonTerminalTasks() ([]*types.EremeticTask, error)
}

func NewDB(driver string, location string) (TaskDB, error) {
	return boltDB(location)
}

func applyMask(task *types.EremeticTask) {
	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = "*******"
	}
}
