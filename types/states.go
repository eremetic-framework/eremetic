package types

// IsTerminal takes a string representation of a state and returns whether it
// is terminal or not.
func IsTerminal(state string) bool {
	switch state {
	case "TASK_LOST", "TASK_KILLED", "TASK_FAILED", "TASK_FINISHED":
		return true
	default:
		return false
	}
}
