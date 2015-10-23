package main

import (
	"fmt"
	"net/http"

	"github.com/alde/eremetic/handler"
	"github.com/alde/eremetic/routes"
	log "github.com/dmuth/google-go-log4go"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

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

func main() {
	readConfig()
	setupLogging()
	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	router := routes.Create()
	log.Infof("listening to %s", bind)
	go handler.Run()
	err := http.ListenAndServe(bind, router)
	log.Info(err.Error())
}
