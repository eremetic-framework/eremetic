package mock

import (
	"errors"

	"github.com/eremetic-framework/eremetic"
)

// Scheduler mocks the eremetic scheduler.
type Scheduler struct {
	ScheduleTaskFn      func(req eremetic.Request) (string, error)
	ScheduleTaskInvoked bool
	KillFn              func(id string) error
	KillInvoked         bool
}

// ScheduleTask invokes the ScheduleTaskFn function.
func (s *Scheduler) ScheduleTask(req eremetic.Request) (string, error) {
	s.ScheduleTaskInvoked = true
	return s.ScheduleTaskFn(req)
}

// Kill simulates the Kill functionality
func (s *Scheduler) Kill(id string) error {
	s.KillInvoked = true
	return s.KillFn(id)
}

// TaskDB mocks the eremetic task database.
type TaskDB struct {
	CleanFn                func() error
	CloseFn                func()
	PutTaskFn              func(*eremetic.Task) error
	ReadTaskFn             func(string) (eremetic.Task, error)
	ReadUnmaskedTaskFn     func(string) (eremetic.Task, error)
	DeleteTaskFn           func(string) error
	ListNonTerminalTasksFn func() ([]*eremetic.Task, error)
	ListTasksFn            func(*eremetic.TaskFilter) ([]*eremetic.Task, error)
}

// Clean invokes the CleanFn function.
func (db *TaskDB) Clean() error {
	return db.CleanFn()
}

// Close invokes the CloseFn function.
func (db *TaskDB) Close() {
	db.CloseFn()
}

// PutTask invokes the PutTaskFn function.
func (db *TaskDB) PutTask(task *eremetic.Task) error {
	return db.PutTaskFn(task)
}

// ReadTask invokes the ReadTaskFn function.
func (db *TaskDB) ReadTask(id string) (eremetic.Task, error) {
	return db.ReadTaskFn(id)
}

// ReadUnmaskedTask invokes the ReadUnmaskedTaskFn function.
func (db *TaskDB) ReadUnmaskedTask(id string) (eremetic.Task, error) {
	return db.ReadUnmaskedTaskFn(id)
}

// DeleteTask invokes the DeleteTaskFn function.
func (db *TaskDB) DeleteTask(id string) error {
	return db.DeleteTaskFn(id)
}

// ListNonTerminalTasks invokes the ListNonTerminalTasksFn function.
func (db *TaskDB) ListNonTerminalTasks() ([]*eremetic.Task, error) {
	return db.ListNonTerminalTasksFn()
}

// ListTasks invokes the ListTasksFn function.
func (db *TaskDB) ListTasks(filter *eremetic.TaskFilter) ([]*eremetic.Task, error) {
	return db.ListTasksFn(filter)
}

// ErrScheduler mocks the eremetic scheduler.
type ErrScheduler struct {
	NextError *error
}

// ScheduleTask records any scheduling errors.
func (s *ErrScheduler) ScheduleTask(request eremetic.Request) (string, error) {
	if err := s.NextError; err != nil {
		s.NextError = nil
		return "", *err

	}
	return "eremetic-task.mock", nil
}

// Kill simulates the Kill functionality
func (s *ErrScheduler) Kill(_id string) error {
	return nil
}

// ErrorReader simulates a failure to read stream.
type ErrorReader struct{}

// Read always returns an error.
func (r *ErrorReader) Read(p []byte) (int, error) {
	return 0, errors.New("oh no")

}
