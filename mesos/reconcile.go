package mesos

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	mesossched "github.com/mesos/mesos-go/api/v0/scheduler"

	"github.com/klarna/eremetic"
)

var (
	maxReconciliationDelay = 120
)

type reconciler struct {
	cancel chan struct{}
	done   chan struct{}
}

func (r *reconciler) Cancel() {
	close(r.cancel)
}

func reconcileTasks(driver mesossched.SchedulerDriver, database eremetic.TaskDB) *reconciler {
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
				ntasks := []*eremetic.Task{}
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
					var statuses []*mesosproto.TaskStatus
					for _, t := range tasks {
						statuses = append(statuses, &mesosproto.TaskStatus{
							State:   mesosproto.TaskState_TASK_STAGING.Enum(),
							TaskId:  &mesosproto.TaskID{Value: proto.String(t.ID)},
							SlaveId: &mesosproto.SlaveID{Value: proto.String(t.SlaveID)},
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

	return &reconciler{
		cancel: cancel,
		done:   done,
	}
}
