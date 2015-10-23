package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/alde/eremetic/handler"
	"github.com/alde/eremetic/routes"
	log "github.com/dmuth/google-go-log4go"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

const version = "0.3.0"

func readConfig() {
	path, _ := osext.ExecutableFolder()
	viper.AddConfigPath("/etc/eremetic")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigName("eremetic")
	viper.SetDefault("loglevel", "debug")
	viper.ReadInConfig()
}

func setupLogging() {
	log.SetLevelString(viper.GetString("loglevel"))
	log.SetDisplayTime(true)
}

func handleFlags() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Print version and exit.")
	flag.Parse()
	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	handleFlags()
	readConfig()
	setupLogging()
	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	// Catch interrupt
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		if s != os.Interrupt {
			return
		}

		log.Info("Eremetic is shutting down")
	}()

	router := routes.Create()
	log.Infof("listening to %s", bind)
	go handler.Run()
	go handler.CleanupTasks()
	err := http.ListenAndServe(bind, router)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
