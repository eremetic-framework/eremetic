package eremetic

import (
	"encoding/json"
	"errors"
	"sync"
)

// Masking is the string used for masking environment variables.
const Masking = "*******"

// ApplyMask replaces masked environment variables with a masking string.
func ApplyMask(task *Task) {
	for k := range task.MaskedEnvironment {
		task.MaskedEnvironment[k] = Masking
	}
}

// Encode encodes a task into a JSON byte array.
func Encode(task *Task) ([]byte, error) {
	encoded, err := json.Marshal(task)
	return []byte(encoded), err
}

// TaskDB defines the functions needed by the database abstraction layer
type TaskDB interface {
	Clean() error
	Close()
	PutTask(task *Task) error
	ReadTask(id string) (Task, error)
	ReadUnmaskedTask(id string) (Task, error)
	ListNonTerminalTasks() ([]*Task, error)
}

// DefaultTaskDB is a in-memory implementation of TaskDB.
type DefaultTaskDB struct {
	mtx   sync.RWMutex
	tasks map[string]*Task
}

// NewDefaultTaskDB returns a new instance of TaskDB.
func NewDefaultTaskDB() *DefaultTaskDB {
	return &DefaultTaskDB{
		tasks: make(map[string]*Task),
	}
}

// Clean removes all tasks from the database.
func (db *DefaultTaskDB) Clean() error {
	db.tasks = make(map[string]*Task)
	return nil
}

// Close closes the connection to the database.
func (db *DefaultTaskDB) Close() {
	return
}

// PutTask adds a new task to the database.
func (db *DefaultTaskDB) PutTask(task *Task) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	db.tasks[task.ID] = task
	return nil
}

// ReadTask returns a task with a given id, or an error if not found.
func (db *DefaultTaskDB) ReadTask(id string) (Task, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	if task, ok := db.tasks[id]; ok {
		ApplyMask(task)
		return *task, nil
	}
	return Task{}, errors.New("unknown task")
}

// ReadUnmaskedTask returns a task with all its environment variables unmasked.
func (db *DefaultTaskDB) ReadUnmaskedTask(id string) (Task, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	if task, ok := db.tasks[id]; ok {
		return *task, nil
	}
	return Task{}, errors.New("unknown task")
}

// ListNonTerminalTasks returns all non-terminal tasks.
func (db *DefaultTaskDB) ListNonTerminalTasks() ([]*Task, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	res := []*Task{}
	for _, t := range db.tasks {
		if !t.IsTerminated() {
			res = append(res, t)
		}
	}
	return res, nil
}
