package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/routes"
	"github.com/klarna/eremetic/scheduler"
	"github.com/prometheus/client_golang/prometheus"
)

func setup() *config.Config {
	cfg := config.DefaultConfig(Version, BuildDate)
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
	prometheus.MustRegister(scheduler.TasksCreated)
	prometheus.MustRegister(scheduler.TasksLaunched)
	prometheus.MustRegister(scheduler.TasksTerminated)
	prometheus.MustRegister(scheduler.TasksDelayed)
	prometheus.MustRegister(scheduler.TasksRunning)
	prometheus.MustRegister(scheduler.QueueSize)
}

func getSchedulerSettings(config *config.Config) *scheduler.Settings {
	return &scheduler.Settings{
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
		fmt.Println(Version)
		os.Exit(0)
	}
	config := setup()

	setupLogging(config.LogFormat, config.LogLevel)
	setupMetrics()
	db, err := database.NewDB(config.DatabaseDriver, config.DatabasePath)

	if err != nil {
		logrus.WithError(err).Fatal("Unable to set up database.")
	}
	defer db.Close()

	schedulerSettings := getSchedulerSettings(config)
	sched := scheduler.Create(schedulerSettings, db)
	go func() {
		scheduler.Run(sched, schedulerSettings)
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

	router := routes.Create(sched, config, db)

	bind := fmt.Sprintf("%s:%d", config.Address, config.Port)
	logrus.WithFields(logrus.Fields{
		"version": config.Version,
		"address": config.Address,
		"port":    config.Port,
	}).Infof("Launching Eremetic version %s!\nListening to %s", config.Version, bind)
	err = manners.ListenAndServe(bind, router)

	if err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error")
	}
}
