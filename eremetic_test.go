package main

import (
	"testing"

	log "github.com/dmuth/google-go-log4go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestMain(t *testing.T) {
	Convey("readConfig", t, func() {
		readConfig()
		So(viper.AllKeys(), ShouldContain, "loglevel")
	})

	Convey("setupLogging", t, func() {
		setupLogging()
		So(log.DisplayTime(), ShouldBeTrue)
		So(log.Level(), ShouldEqual, log.DebugLevel)
	})
}
