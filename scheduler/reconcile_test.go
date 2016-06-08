package scheduler

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestReconcile(t *testing.T) {
	dir, _ := os.Getwd()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil {
		t.Fail()
	}

	db.Clean()
	defer db.Close()

	maxReconciliationDelay = 1

	Convey("ReconcileTasks", t, func() {
		Convey("Finishes when there are no tasks", func() {
			driver := NewMockScheduler()
			r := ReconcileTasks(driver, db)

			select {
			case <-r.done:
			}

			So(driver.AssertNotCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})

		Convey("Sends reconcile request", func() {
			driver := NewMockScheduler()
			driver.On("ReconcileTasks").Run(func(mock.Arguments) {
				t, err := db.ReadTask("1234")
				if err != nil {
					panic("mock error")
				}
				t.UpdateStatus(types.Status{
					Status: mesos.TaskState_TASK_RUNNING.String(),
					Time:   time.Now().Unix() + 1,
				})
				db.PutTask(&t)
			}).Once()

			db.PutTask(&types.EremeticTask{
				ID: "1234",
				Status: []types.Status{
					types.Status{
						Status: mesos.TaskState_TASK_STAGING.String(),
						Time:   time.Now().Unix(),
					},
				},
			})

			r := ReconcileTasks(driver, db)

			select {
			case <-r.done:
			}

			So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})

		Convey("Cancel reconciliation", func() {
			driver := NewMockScheduler()

			db.PutTask(&types.EremeticTask{
				ID: "1234",
				Status: []types.Status{
					types.Status{
						Status: mesos.TaskState_TASK_STAGING.String(),
						Time:   time.Now().Unix(),
					},
				},
			})

			r := ReconcileTasks(driver, db)
			r.Cancel()

			select {
			case <-r.done:
			}

			So(driver.AssertNotCalled(t, "ReconcileTasks"), ShouldBeTrue)
		})
	})
}
