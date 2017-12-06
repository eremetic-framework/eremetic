package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// TasksCreated increments with each created task
	TasksCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_created",
		Help:      "Number of tasks submitted to eremetic",
	})
	// TasksLaunched increments with each launched task
	TasksLaunched = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_launched",
		Help:      "Number of tasks launched by eremetic",
	})
	// TasksTerminated increments with each terminated task
	TasksTerminated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_terminated",
		Help:      "Number of terminated tasks by terminal status",
	}, []string{"status", "sequence"})
	// TasksDelayed increments with each delayed task
	TasksDelayed = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_delayed",
		Help:      "Number of times the launch of a task has been delayed",
	})
	// TasksRunning provides the number of currently running tasks
	TasksRunning = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "scheduler",
		Name:      "tasks_running",
		Help:      "Number of tasks currently running",
	})
	// QueueSize provides the number of tasks waiting to be launched
	QueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "scheduler",
		Name:      "queue_size",
		Help:      "Number of tasks in the queue",
	})
)

// RegisterMetrics registers mesos metrics to a prometheus Registerer.
func RegisterMetrics(r prometheus.Registerer) error {
	errs := []error{
		r.Register(TasksCreated),
		r.Register(TasksLaunched),
		r.Register(TasksTerminated),
		r.Register(TasksDelayed),
		r.Register(TasksRunning),
		r.Register(QueueSize),
	}
	if len(errs) > 0 {
		return errors.New("unable to register metrics")
	}
	return nil
}
