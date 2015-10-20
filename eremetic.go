package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alde/eremetic/routes"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

func readConfig() {
	path, _ := osext.ExecutableFolder()
	viper.AddConfigPath("/etc/eremetic")
	viper.AddConfigPath(path)
	viper.SetConfigName("eremetic")
	viper.ReadInConfig()
}

func main() {
	readConfig()
	bind := fmt.Sprintf("%s:%d", viper.GetString("address"), viper.GetInt("port"))

	router := routes.Create()
	log.Printf("listening to %s", bind)
	log.Fatal(http.ListenAndServe(bind, router))
}
