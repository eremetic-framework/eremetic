package main

import (
	"testing"

	"github.com/Sirupsen/logrus"
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
		So(logrus.GetLevel(), ShouldEqual, logrus.DebugLevel)
	})
}
