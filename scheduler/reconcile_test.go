package scheduler

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/boltdb"
)

func TestReconcile(t *testing.T) {
	dir, _ := os.Getwd()
	db, err := boltdb.NewTaskDB(fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil {
		t.Fatal(err)
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
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskState_TASK_RUNNING,
					Time:   time.Now().Unix() + 1,
				})
				db.PutTask(&t)
			}).Once()

			db.PutTask(&eremetic.Task{
				ID: "1234",
				Status: []eremetic.Status{
					eremetic.Status{
						Status: eremetic.TaskState_TASK_STAGING,
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

			db.PutTask(&eremetic.Task{
				ID: "1234",
				Status: []eremetic.Status{
					eremetic.Status{
						Status: eremetic.TaskState_TASK_STAGING,
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
