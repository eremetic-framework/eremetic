package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	"github.com/kardianos/osext"
	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/janitor"
	"github.com/klarna/eremetic/routes"
	"github.com/klarna/eremetic/scheduler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func readConfig() {
	path, _ := osext.ExecutableFolder()
	viper.AddConfigPath("/etc/eremetic")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigName("eremetic")
	viper.SetDefault("name", "Eremetic")
	viper.SetDefault("user", "root")
	viper.SetDefault("loglevel", "debug")
	viper.SetDefault("logformat", "text")
	viper.SetDefault("database_driver", "boltdb")
	viper.SetDefault("checkpoint", "true")
	viper.SetDefault("failover_timeout", 2592000.0)
	viper.SetDefault("queue_size", 100)
	viper.SetDefault("janitor", "false")
	viper.SetDefault("janitor_retention_period", 0)
	viper.SetDefault("janitor_pause", 300)
	viper.ReadInConfig()

	driver := viper.GetString("database_driver")
	location := viper.GetString("database")
	if driver == "zk" && location == "" {
		viper.Set("database", "")
	} else if driver == "boltdb" && location == "" {
		viper.Set("database", "db/eremetic.db")
	}
}

func setupLogging() {
	if viper.GetString("logformat") == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	level, err := logrus.ParseLevel(viper.GetString("loglevel"))
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

func getSchedulerSettings() *scheduler.Settings {
	return &scheduler.Settings{
		MaxQueueSize:     viper.GetInt("queue_size"),
		Master:           viper.GetString("master"),
		FrameworkID:      viper.GetString("framework_id"),
		CredentialFile:   viper.GetString("credential_file"),
		Name:             viper.GetString("name"),
		User:             viper.GetString("user"),
		MessengerAddress: viper.GetString("messenger_address"),
		MessengerPort:    uint16(viper.GetInt("messenger_port")),
		Checkpoint:       viper.GetBool("checkpoint"),
		FailoverTimeout:  viper.GetFloat64("failover_timeout"),
	}
}

func setupDB() (database.TaskDB, error) {
	driver := viper.GetString("database_driver")
	location := viper.GetString("database")
	return database.NewDB(driver, location)
}

func setupJanitor(db database.TaskDB) janitor.Janitor {
	enabled := viper.GetBool("janitor")
	retentionPeriod := viper.GetInt64("janitor_retention_period")
	pause := viper.GetInt64("janitor_pause")
	return janitor.NewJanitor(db, enabled, retentionPeriod, pause)
}
func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(Version)
		os.Exit(0)
	}
	readConfig()
	setupLogging()
	setupMetrics()
	db, err := setupDB()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to set up database.")
	}
	defer db.Close()

	config := &config.Config{
		Version:   strings.Trim(Version, "'"),
		BuildDate: BuildDate,
		Database:  db,
	}

	schedulerSettings := getSchedulerSettings()

	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	sched := scheduler.Create(schedulerSettings, config)
	go func() {
		scheduler.Run(sched, schedulerSettings)
		manners.Close()
	}()

	// The GrimReaper
	grimreaper := setupJanitor(db)
	go func() {
		grimreaper.Run()
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

	router := routes.Create(sched, config)
	logrus.WithFields(logrus.Fields{
		"version": config.Version,
		"address": viper.GetString("address"),
		"port":    viper.GetInt("port"),
	}).Infof("Launching Eremetic version %s!\nListening to %s", config.Version, bind)
	err = manners.ListenAndServe(bind, router)

	if err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error")
	}
}
