package main

import (
	"testing"

	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestMain(t *testing.T) {
	Convey("readConfig", t, func() {
		Convey("Defaults", func() {
			readConfig()
			keys := viper.AllKeys()
			So(keys, ShouldContain, "name")
			So(keys, ShouldContain, "user")
			So(keys, ShouldContain, "loglevel")
			So(keys, ShouldContain, "logformat")
			So(keys, ShouldContain, "database")
			So(keys, ShouldContain, "checkpoint")
			So(keys, ShouldContain, "failover_timeout")
			So(keys, ShouldContain, "queue_size")
		})
	})

	Convey("setupLogging", t, func() {
		setupLogging()
		So(logrus.GetLevel(), ShouldEqual, logrus.DebugLevel)
	})
}
