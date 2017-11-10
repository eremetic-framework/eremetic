package mesos

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler/calls"

	"github.com/eremetic-framework/eremetic"
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

func (s *Scheduler) reconcileTasks() *reconciler {
	frameworkOpt := calls.Framework(s.frameworkID)
	cancel := make(chan struct{})
	done := make(chan struct{})

	go func() {
		var (
			c     uint
			delay = 1
		)

		tasks, err := s.database.ListNonTerminalTasks()
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
					nt, err := s.database.ReadTask(t.ID)
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
					taskmap := make(map[string]string)
					for _, t := range tasks {
						taskmap[t.ID] = t.AgentID
					}
					logrus.WithField("reconciliation_request_count", c).Debug("Sending reconciliation request")
					reconcile := calls.Reconcile(calls.ReconcileTasks(taskmap)).With(frameworkOpt)
					if err := calls.CallNoData(s.caller, reconcile); err != nil {
						logrus.WithError(err).Warn("Failed to send reconciliation request")
					}
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
