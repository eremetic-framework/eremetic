package scheduler

import (
	log "github.com/dmuth/google-go-log4go"
)

func Create() *eremeticScheduler {
	return createEremeticScheduler()
}

// Run the eremetic scheduler
func Run(s *eremeticScheduler) {
	driver, err := createDriver(s)

	if err != nil {
		log.Errorf("Unable to create scheduler driver: %s", err)
		return
	}

	defer close(s.shutdown)
	defer driver.Stop(false)

	if status, err := driver.Run(); err != nil {
		log.Errorf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
	}
	log.Info("Exiting...")
}
