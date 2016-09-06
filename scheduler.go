package eremetic

type Scheduler interface {
	ScheduleTask(request Request) (string, error)
}
