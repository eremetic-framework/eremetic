package eremetic

import (
	"encoding/json"
	"errors"
	"sync"
)

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

// TaskDB defines the functions needed by the database abstraction layer
type TaskDB interface {
	Clean() error
	Close()
	PutTask(task *Task) error
	ReadTask(id string) (Task, error)
	ReadUnmaskedTask(id string) (Task, error)
	ListNonTerminalTasks() ([]*Task, error)
}

type DefaultTaskDB struct {
	mtx   sync.RWMutex
	tasks map[string]*Task
}

func NewDefaultTaskDB() *DefaultTaskDB {
	return &DefaultTaskDB{
		tasks: make(map[string]*Task),
	}
}

func (db *DefaultTaskDB) Clean() error {
	db.tasks = make(map[string]*Task)
	return nil
}

func (db *DefaultTaskDB) Close() {
	return
}

func (db *DefaultTaskDB) PutTask(task *Task) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	db.tasks[task.ID] = task
	return nil
}

func (db *DefaultTaskDB) ReadTask(id string) (Task, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	if task, ok := db.tasks[id]; ok {
		ApplyMask(task)
		return *task, nil
	}
	return Task{}, errors.New("unknown cargo")
}

func (db *DefaultTaskDB) ReadUnmaskedTask(id string) (Task, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	if task, ok := db.tasks[id]; ok {
		return *task, nil
	}
	return Task{}, errors.New("unknown cargo")
}

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
