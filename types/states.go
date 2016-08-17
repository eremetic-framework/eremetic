package types

type TaskState string

const (
	TaskState_TASK_STAGING  TaskState = "TASK_STAGING"
	TaskState_TASK_STARTING TaskState = "TASK_STARTING"
	TaskState_TASK_RUNNING  TaskState = "TASK_RUNNING"
	TaskState_TASK_FINISHED TaskState = "TASK_FINISHED"
	TaskState_TASK_FAILED   TaskState = "TASK_FAILED"
	TaskState_TASK_KILLED   TaskState = "TASK_KILLED"
	TaskState_TASK_LOST     TaskState = "TASK_LOST"
	TaskState_TASK_ERROR    TaskState = "TASK_ERROR"
)

// IsTerminal takes a string representation of a state and returns whether it
// is terminal or not.
func IsTerminal(state TaskState) bool {
	switch state {
	case "TASK_LOST", "TASK_KILLED", "TASK_FAILED", "TASK_FINISHED":
		return true
	default:
		return false
	}
}
