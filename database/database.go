package database

import (
	"encoding/json"
	"errors"

	"github.com/klarna/eremetic/types"
)

// TaskDB defines the functions needed by the database abstraction layer
type TaskDB interface {
	Clean() error
	Close()
	Count() int
	PutTask(task *types.EremeticTask) error
	ReadTask(id string) (types.EremeticTask, error)
	ReadUnmaskedTask(id string) (types.EremeticTask, error)
	ListNonTerminalTasks() ([]*types.EremeticTask, error)
}

const masking = "*******"

// NewDB Is used to create a new database driver based on settings.
func NewDB(driver string, location string) (TaskDB, error) {
	switch driver {
	case "boltdb":
		return createBoltDriver(createBoltConnector(), location)
	case "zk":
		return createZKDriver(createZKConnector(), location)
	}
	return nil, errors.New("Invalid driver.")
}

func applyMask(task *types.EremeticTask) {
	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = masking
	}
}

func encode(task *types.EremeticTask) ([]byte, error) {
	encoded, err := json.Marshal(task)
	return []byte(encoded), err
}
