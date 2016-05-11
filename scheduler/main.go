package scheduler

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// Create is used to build a new scheduler with default settings
func Create() *eremeticScheduler {
	return createEremeticScheduler(&Settings{
		MaxQueueSize: viper.GetInt("queue_size"),
	})
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
