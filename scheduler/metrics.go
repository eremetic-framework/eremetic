package scheduler

import "github.com/prometheus/client_golang/prometheus"

var (
	TasksCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_created",
		Help:      "Number of tasks submitted to eremetic",
	})
	TasksLaunched = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_launched",
		Help:      "Number of tasks launched by eremetic",
	})
	TasksTerminated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_terminated",
		Help:      "Number of terminated tasks by terminal status",
	}, []string{"status"})
	TasksDelayed = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "scheduler",
		Name:      "tasks_delayed",
		Help:      "Number of times the launch of a task has been delayed",
	})
	TasksRunning = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "scheduler",
		Name:      "tasks_running",
		Help:      "Number of tasks currently running",
	})
	QueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "scheduler",
		Name:      "queue_size",
		Help:      "Number of tasks in the queue",
	})
)
