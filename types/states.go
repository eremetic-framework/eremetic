package types

import mesos "github.com/mesos/mesos-go/mesosproto"

// IsTerminal determines if a TaskState is a terminal state, i.e. if it singals
// that the task has stopped running.
func IsTerminal(state *mesos.TaskState) bool {
	switch *state {
	case mesos.TaskState_TASK_FINISHED,
		mesos.TaskState_TASK_FAILED,
		mesos.TaskState_TASK_KILLED,
		mesos.TaskState_TASK_LOST:
		return true
	default:
		return false
	}
}

// IsTerminalString takes a string representation of a state and returns whether
// it is terminal or not.
func IsTerminalString(state string) bool {
	switch state {
	case "TASK_LOST", "TASK_KILLED", "TASK_FAILED", "TASK_FINISHED":
		return true
	default:
		return false
	}
}
