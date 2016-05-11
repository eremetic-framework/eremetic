package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	"github.com/kardianos/osext"
	"github.com/klarna/eremetic/database"
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
	viper.SetDefault("database", "db/eremetic.db")
	viper.SetDefault("checkpoint", "true")
	viper.SetDefault("failover_timeout", 2592000.0)
	viper.SetDefault("queue_size", 100)
	viper.ReadInConfig()
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

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(Version)
		os.Exit(0)
	}
	readConfig()
	setupLogging()
	setupMetrics()
	defer database.Close()

	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	sched := scheduler.Create()
	go func() {
		scheduler.Run(sched)
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

	router := routes.Create(sched)
	logrus.WithFields(logrus.Fields{
		"address": viper.GetString("address"),
		"port":    viper.GetInt("port"),
	}).Infof("listening to %s", bind)
	err := manners.ListenAndServe(bind, router)
	database.Close()

	if err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error")
	}
}
