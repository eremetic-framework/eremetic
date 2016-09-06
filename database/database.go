package database

import (
	"encoding/json"
	"errors"

	"github.com/klarna/eremetic"
)

// TaskDB defines the functions needed by the database abstraction layer
type TaskDB interface {
	Clean() error
	Close()
	PutTask(task *eremetic.Task) error
	ReadTask(id string) (eremetic.Task, error)
	ReadUnmaskedTask(id string) (eremetic.Task, error)
	ListNonTerminalTasks() ([]*eremetic.Task, error)
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

func applyMask(task *eremetic.Task) {
	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = masking
	}
}

func encode(task *eremetic.Task) ([]byte, error) {
	encoded, err := json.Marshal(task)
	return []byte(encoded), err
}
