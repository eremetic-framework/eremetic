package janitor

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/klarna/eremetic/database"
)

// Janitor
type Janitor struct {
	db              database.TaskDB
	enabled         bool
	retentionPeriod int64
	pause           int64
}

func NewJanitor(db database.TaskDB, enabled bool, retentionPeriod int64, pause int64) Janitor {
	return Janitor{
		db:              db,
		enabled:         enabled,
		retentionPeriod: retentionPeriod,
		pause:           pause,
	}
}

func (j Janitor) Run() {
	if j.enabled {
		j.runLoop()
	} else {
		j.runWaitForEvent()
	}
}

func (j Janitor) runWaitForEvent() {
	select {}
}

func (j Janitor) runLoop() {
	j.purgeTerminatedTasks()
	time.Sleep(time.Duration(j.pause) * time.Second)
}

func (j Janitor) purgeTerminatedTasks() error {
	tasks, err := j.db.ListTerminatedTasks()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.IsExpired(j.retentionPeriod) {
			err = j.db.RemoveTask(t.ID)
			if err != nil {
				logrus.WithError(err).Error("Purge of tasks failed. Stopping purge process.")
				break
			}
		}
	}
	return err
}
