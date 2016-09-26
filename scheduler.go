package eremetic

import "errors"

// ErrQueueFull is returned in the event of a full queue. This allows the caller
// to handle this as they see fit.
var ErrQueueFull = errors.New("task queue is full")

// Scheduler defines an interface for scheduling tasks.
type Scheduler interface {
	ScheduleTask(request Request) (string, error)
}
