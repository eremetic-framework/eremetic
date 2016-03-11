package scheduler

import (
	"github.com/Sirupsen/logrus"
)

// Create is used to build a new scheduler
func Create() *eremeticScheduler {
	return createEremeticScheduler()
}

// Run the eremetic scheduler
func Run(s *eremeticScheduler) {
	driver, err := createDriver(s)

	if err != nil {
		logrus.WithError(err).Error("Unable to create scheduler driver")
		return
	}

	go func() {
		<-s.shutdown
		driver.Stop(false)
	}()

	if status, err := driver.Run(); err != nil {
		logrus.WithError(err).WithField("status", status.String()).Error("Framework stopped")
	}
	logrus.Info("Exiting...")
}
