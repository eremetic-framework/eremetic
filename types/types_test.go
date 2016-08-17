package types

import (
	"fmt"
	"testing"

	mesos "github.com/mesos/mesos-go/mesosproto"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypes(t *testing.T) {
	Convey("states", t, func() {
		terminalStates := []mesos.TaskState{
			mesos.TaskState_TASK_FINISHED,
			mesos.TaskState_TASK_FAILED,
			mesos.TaskState_TASK_KILLED,
			mesos.TaskState_TASK_LOST,
		}

		nonTerminalStates := []mesos.TaskState{
			mesos.TaskState_TASK_RUNNING,
			mesos.TaskState_TASK_STAGING,
			mesos.TaskState_TASK_STARTING,
		}

		Convey("IsTerminal", func() {
			for _, state := range terminalStates {
				test := fmt.Sprintf("Should be true for %s", state.String())
				Convey(test, func() {
					So(IsTerminal(state.String()), ShouldBeTrue)
				})
			}

			for _, state := range nonTerminalStates {
				test := fmt.Sprintf("Should be false for %s", state.String())
				Convey(test, func() {
					So(IsTerminal(state.String()), ShouldBeFalse)
				})
			}
		})
	})
}
