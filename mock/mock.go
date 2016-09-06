package mock

import "github.com/klarna/eremetic"

type Scheduler struct {
	ScheduleTaskFn      func(req eremetic.Request) (string, error)
	ScheduleTaskInvoked bool
}

func (s *Scheduler) ScheduleTask(req eremetic.Request) (string, error) {
	s.ScheduleTaskInvoked = true
	return s.ScheduleTaskFn(req)
}

type TaskDB struct {
	CleanFn                func() error
	CloseFn                func()
	PutTaskFn              func(*eremetic.Task) error
	ReadTaskFn             func(string) (eremetic.Task, error)
	ReadUnmaskedTaskFn     func(string) (eremetic.Task, error)
	ListNonTerminalTasksFn func() ([]*eremetic.Task, error)
}

func (db *TaskDB) Clean() error {
	return db.CleanFn()
}

func (db *TaskDB) Close() {
	db.CloseFn()
}

func (db *TaskDB) PutTask(task *eremetic.Task) error {
	return db.PutTaskFn(task)
}

func (db *TaskDB) ReadTask(id string) (eremetic.Task, error) {
	return db.ReadTaskFn(id)
}

func (db *TaskDB) ReadUnmaskedTask(id string) (eremetic.Task, error) {
	return db.ReadUnmaskedTaskFn(id)
}

func (db *TaskDB) ListNonTerminalTasks() ([]*eremetic.Task, error) {
	return db.ListNonTerminalTasksFn()
}
