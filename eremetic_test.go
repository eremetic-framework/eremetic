package main

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/klarna/eremetic/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(t *testing.T) {
	conf := config.DefaultConfig()
	Convey("GetSchedulerSettings", t, func() {
		Convey("Contains defaults", func() {
			s := getSchedulerSettings(conf)

			So(s.MaxQueueSize, ShouldEqual, 100)
			So(s.Master, ShouldEqual, "")
			So(s.FrameworkID, ShouldEqual, "")
			So(s.CredentialFile, ShouldEqual, "")
			So(s.Name, ShouldEqual, "Eremetic")
			So(s.User, ShouldEqual, "root")
			So(s.MessengerAddress, ShouldEqual, "")
			So(s.MessengerPort, ShouldEqual, 0)
			So(s.Checkpoint, ShouldEqual, true)
			So(s.FailoverTimeout, ShouldAlmostEqual, 2592000.0)
		})
	})

	Convey("setupLogging", t, func() {
		setupLogging(conf.LogFormat, conf.LogLevel)
		So(logrus.GetLevel(), ShouldEqual, logrus.DebugLevel)
	})
}
