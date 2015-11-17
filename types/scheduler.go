package types

type Scheduler interface {
	ScheduleTask(request Request) (string, error)
}
