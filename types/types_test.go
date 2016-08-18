package types

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypes(t *testing.T) {
	Convey("states", t, func() {
		terminalStates := []TaskState{
			TaskState_TASK_FINISHED,
			TaskState_TASK_FAILED,
			TaskState_TASK_KILLED,
			TaskState_TASK_LOST,
		}

		nonTerminalStates := []TaskState{
			TaskState_TASK_RUNNING,
			TaskState_TASK_STAGING,
			TaskState_TASK_STARTING,
		}

		Convey("IsTerminal", func() {
			for _, state := range terminalStates {
				test := fmt.Sprintf("Should be true for %s", state)
				Convey(test, func() {
					So(IsTerminal(state), ShouldBeTrue)
				})
			}

			for _, state := range nonTerminalStates {
				test := fmt.Sprintf("Should be false for %s", state)
				Convey(test, func() {
					So(IsTerminal(state), ShouldBeFalse)
				})
			}
		})
	})
}
