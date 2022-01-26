package mesos

import (
	"testing"
	"time"

	"github.com/mesos/mesos-go/api/v0/mesosproto"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/rockerbox/eremetic"
	"github.com/rockerbox/eremetic/mock"
)

func TestReconcile(t *testing.T) {
	db := eremetic.NewDefaultTaskDB()

	maxReconciliationDelay = 1

	Convey("ReconcileTasks", t, func() {
		Convey("Finishes when there are no tasks", func() {
			driver := mock.NewMesosScheduler()
			r := reconcileTasks(driver, db)

			select {
			case <-r.done:
			}

			So(driver.ReconcileTasksFnInvoked, ShouldBeFalse)
		})

		Convey("Sends reconcile request", func() {
			driver := mock.NewMesosScheduler()
			driver.ReconcileTasksFn = func(ts []*mesosproto.TaskStatus) (mesosproto.Status, error) {
				t, err := db.ReadTask("1234")
				if err != nil {
					return mesosproto.Status_DRIVER_RUNNING, err
				}
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskRunning,
					Time:   time.Now().Unix() + 1,
				})
				db.PutTask(&t)

				return mesosproto.Status_DRIVER_RUNNING, nil
			}

			db.PutTask(&eremetic.Task{
				ID: "1234",
				Status: []eremetic.Status{
					eremetic.Status{
						Status: eremetic.TaskStaging,
						Time:   time.Now().Unix(),
					},
				},
			})

			r := reconcileTasks(driver, db)

			select {
			case <-r.done:
			}

			So(driver.ReconcileTasksFnInvoked, ShouldBeTrue)
		})

		Convey("Cancel reconciliation", func() {
			driver := mock.NewMesosScheduler()

			db.PutTask(&eremetic.Task{
				ID: "1234",
				Status: []eremetic.Status{
					eremetic.Status{
						Status: eremetic.TaskStaging,
						Time:   time.Now().Unix(),
					},
				},
			})

			r := reconcileTasks(driver, db)
			r.Cancel()

			select {
			case <-r.done:
			}

			So(driver.ReconcileTasksFnInvoked, ShouldBeFalse)
		})
	})
}
