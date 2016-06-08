package scheduler

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
)

var (
	maxReconciliationDelay = 120
)

type Reconcile struct {
	cancel chan struct{}
	done   chan struct{}
}

func (r *Reconcile) Cancel() {
	close(r.cancel)
}

func ReconcileTasks(driver sched.SchedulerDriver, database database.TaskDB) *Reconcile {
	cancel := make(chan struct{})
	done := make(chan struct{})

	go func() {
		var (
			c     uint
			delay = 1
		)

		tasks, err := database.ListNonTerminalTasks()
		if err != nil {
			logrus.WithError(err).Error("Failed to list non-terminal tasks")
			close(done)
			return
		}

		logrus.Infof("Trying to reconcile with %d task(s)", len(tasks))
		start := time.Now()

		for len(tasks) > 0 {
			select {
			case <-cancel:
				logrus.Info("Cancelling reconciliation job")
				close(done)
				return
			case <-time.After(time.Duration(delay) * time.Second):
				// Filter tasks that has received a status update
				ntasks := []*types.EremeticTask{}
				for _, t := range tasks {
					nt, err := database.ReadTask(t.ID)
					if err != nil {
						logrus.WithField("task_id", t.ID).Warn("Task not found in database")
						continue
					}
					if nt.LastUpdated().Before(start) {
						ntasks = append(ntasks, &nt)
					}
				}
				tasks = ntasks

				// Send reconciliation request
				if len(tasks) > 0 {
					var statuses []*mesos.TaskStatus
					for _, t := range tasks {
						statuses = append(statuses, &mesos.TaskStatus{
							State:   mesos.TaskState_TASK_STAGING.Enum(),
							TaskId:  &mesos.TaskID{Value: proto.String(t.ID)},
							SlaveId: &mesos.SlaveID{Value: proto.String(t.SlaveId)},
						})
					}
					logrus.WithField("reconciliation_request_count", c).Debug("Sending reconciliation request")
					driver.ReconcileTasks(statuses)
				}

				if delay < maxReconciliationDelay {
					delay = 10 << c
					if delay >= maxReconciliationDelay {
						delay = maxReconciliationDelay
					}
				}

				c++
			}
		}

		logrus.Info("Reconciliation done")
		close(done)
	}()

	return &Reconcile{
		cancel: cancel,
		done:   done,
	}
}
