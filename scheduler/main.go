package scheduler

import (
	"github.com/Sirupsen/logrus"
)

// Run the eremetic scheduler
func Run(s *eremeticScheduler, settings *Settings) {
	driver, err := createDriver(s, settings)

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
