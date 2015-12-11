package scheduler

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alde/eremetic/database"
	"github.com/alde/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestReconcile(t *testing.T) {
	dir, _ := os.Getwd()
	database.NewDB(fmt.Sprintf("%s/../db/test.db", dir))
	database.Clean()
	defer database.Close()

	maxReconciliationDelay = 1

	Convey("ReconcileTasks", t, func() {
		Convey("Finishes when there are no tasks", func() {
			driver := NewMockScheduler()
			r := ReconcileTasks(driver)

			select {
			case <-r.done:
			}

			So(driver.AssertNotCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})

		Convey("Sends reconcile request", func() {
			driver := NewMockScheduler()
			driver.On("ReconcileTasks").Run(func(mock.Arguments) {
				t, err := database.ReadTask("1234")
				if err != nil {
					panic("mock error")
				}
				t.UpdateStatus(types.Status{
					Status: mesos.TaskState_TASK_RUNNING.String(),
					Time:   time.Now().Unix() + 1,
				})
				database.PutTask(&t)
			}).Once()

			database.PutTask(&types.EremeticTask{
				ID: "1234",
				Status: []types.Status{
					types.Status{
						Status: mesos.TaskState_TASK_STAGING.String(),
						Time:   time.Now().Unix(),
					},
				},
			})

			r := ReconcileTasks(driver)

			select {
			case <-r.done:
			}

			So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})

		Convey("Cancel reconciliation", func() {
			driver := NewMockScheduler()

			database.PutTask(&types.EremeticTask{
				ID: "1234",
				Status: []types.Status{
					types.Status{
						Status: mesos.TaskState_TASK_STAGING.String(),
						Time:   time.Now().Unix(),
					},
				},
			})

			r := ReconcileTasks(driver)
			r.Cancel()

			select {
			case <-r.done:
			}

			So(driver.AssertNotCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})
	})
}
