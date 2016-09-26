package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/boltdb"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/mesos"
	"github.com/klarna/eremetic/server"
	"github.com/klarna/eremetic/version"
	"github.com/klarna/eremetic/zk"
)

func setup() *config.Config {
	cfg := config.DefaultConfig()
	config.ReadConfigFile(cfg, config.GetConfigFilePath())
	config.ReadEnvironment(cfg)

	return cfg
}

func setupLogging(logFormat, logLevel string) {
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

func setupMetrics() {
	prometheus.MustRegister(mesos.TasksCreated)
	prometheus.MustRegister(mesos.TasksLaunched)
	prometheus.MustRegister(mesos.TasksTerminated)
	prometheus.MustRegister(mesos.TasksDelayed)
	prometheus.MustRegister(mesos.TasksRunning)
	prometheus.MustRegister(mesos.QueueSize)
}

func getSchedulerSettings(config *config.Config) *mesos.Settings {
	return &mesos.Settings{
		MaxQueueSize:     config.QueueSize,
		Master:           config.Master,
		FrameworkID:      config.FrameworkID,
		CredentialFile:   config.CredentialsFile,
		Name:             config.Name,
		User:             config.User,
		MessengerAddress: config.MessengerAddress,
		MessengerPort:    uint16(config.MessengerPort),
		Checkpoint:       config.Checkpoint,
		FailoverTimeout:  config.FailoverTimeout,
	}
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version.Version)
		os.Exit(0)
	}
	config := setup()

	setupLogging(config.LogFormat, config.LogLevel)
	setupMetrics()

	db, err := NewDB(config.DatabaseDriver, config.DatabasePath)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to set up database.")
	}
	defer db.Close()

	sched := mesos.NewScheduler(config.QueueSize, db)

	go func() {
		sched.Run(getSchedulerSettings(config))
		manners.Close()
	}()

	// Catch interrupt
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		if s != os.Interrupt && s != os.Kill {
			return
		}

		logrus.Info("Eremetic is shutting down")
		sched.Stop()
	}()

	router := server.NewRouter(sched, config, db)

	bind := fmt.Sprintf("%s:%d", config.Address, config.Port)

	logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"address": config.Address,
		"port":    config.Port,
	}).Infof("Launching Eremetic version %s!\nListening to %s", version.Version, bind)

	err = manners.ListenAndServe(bind, router)
	if err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error")
	}
}

// NewDB Is used to create a new database driver based on settings.
func NewDB(driver string, location string) (eremetic.TaskDB, error) {
	switch driver {
	case "boltdb":
		return boltdb.NewTaskDB(location)
	case "zk":
		return zk.NewTaskDB(location)
	}
	return nil, errors.New("Invalid driver.")
}
