package mesos

import (
	"testing"
	"time"

	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/mock"
)

func TestReconcile(t *testing.T) {
	db := eremetic.NewDefaultTaskDB()
	maxReconciliationDelay = 1

	Convey("ReconcileTasks", t, func() {
		Convey("Finishes when there are no tasks", func() {
			caller := mock.NewCaller()
			s := &Scheduler{
				database: db,
				caller:   caller,
			}

			r := s.reconcileTasks()

			select {
			case <-r.done:
			}

			So(caller.CallFnInvoked, ShouldBeFalse)
		})

		Convey("Sends reconcile request", func() {
			caller := mock.NewCaller()
			s := &Scheduler{
				database: db,
				caller:   caller,
			}
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				t, err := db.ReadTask("1234")
				if err != nil {
					return nil, err
				}
				t.UpdateStatus(eremetic.Status{
					Status: eremetic.TaskRunning,
					Time:   time.Now().Unix() + 1,
				})
				db.PutTask(&t)

				return nil, nil
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

			r := s.reconcileTasks()

			select {
			case <-r.done:
			}

			So(caller.CallFnInvoked, ShouldBeTrue)
			So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_RECONCILE)
		})

		Convey("Cancel reconciliation", func() {
			caller := mock.NewCaller()
			s := &Scheduler{
				database: db,
				caller:   caller,
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

			r := s.reconcileTasks()
			r.Cancel()

			select {
			case <-r.done:
			}

			So(caller.CallFnInvoked, ShouldBeFalse)
		})
	})
}
