package main

import (
	"fmt"
	"io/ioutil"
	"os/user"
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

	Convey("GetSchedulerSettings", t, func() {
		Convey("Contains defaults", func() {
			u, err := user.Current()
			So(err, ShouldBeNil)
			s := getSchedulerSettings()
			So(s.MaxQueueSize, ShouldEqual, 100)
			So(s.Master, ShouldEqual, "")
			So(s.FrameworkID, ShouldEqual, "")
			So(s.CredentialFile, ShouldEqual, "")
			So(s.Name, ShouldEqual, "Eremetic")
			So(s.User, ShouldEqual, u.Username)
			So(s.MessengerAddress, ShouldEqual, "")
			So(s.MessengerPort, ShouldEqual, 0)
			So(s.Checkpoint, ShouldEqual, true)
			So(s.FailoverTimeout, ShouldAlmostEqual, 2592000.0)
		})
	})

	Convey("setupLogging", t, func() {
		setupLogging()
		So(logrus.GetLevel(), ShouldEqual, logrus.DebugLevel)
	})

	Convey("setupDB", t, func() {
		Convey("BoltDB", func() {
			dir, _ := ioutil.TempDir("", "eremetic")
			testDB := fmt.Sprintf("%s/test.db", dir)
			viper.Set("database_driver", "boltdb")
			viper.Set("database", testDB)

			db, err := setupDB()

			So(db, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})

		Convey("Invalid configuration", func() {
			viper.Set("database_driver", nil)
			viper.Set("database", nil)

			db, err := setupDB()

			So(db, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
