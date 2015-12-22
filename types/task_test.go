package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("WasRunning", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that never was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.WasRunning(), ShouldBeFalse)
		})
	})

	Convey("IsTerminated", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsTerminated(), ShouldBeFalse)
		})

		Convey("A task that never was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A empty task", func() {
			task := EremeticTask{}

			So(task.IsTerminated(), ShouldBeTrue)
		})
	})

	Convey("IsRunning", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsRunning(), ShouldBeFalse)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsRunning(), ShouldBeTrue)
		})
	})

	Convey("LastUpdated", t, func() {
		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{1449682262, "TASK_STAGING"},
					Status{1449682265, "TASK_RUNNING"},
				},
			}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 1449682265)
		})

		Convey("A empty task", func() {
			task := EremeticTask{}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 0)
		})
	})
}
