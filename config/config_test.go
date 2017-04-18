package config

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	wd, _ := os.Getwd()

	Convey("The Config Builders", t, func() {
		conf := DefaultConfig()

		Convey("ReadConfigFile", func() {
			ReadConfigFile(conf, fmt.Sprintf("%s/config_test.yml", wd))
			So(conf.Port, ShouldEqual, 8080)
			So(conf.Address, ShouldEqual, "0.0.0.0")
			So(conf.HTTPCredentials, ShouldEqual, "admin:admin")
			So(conf.CredentialsFile, ShouldEqual, "/tmp/secret_file")
		})

		Convey("ReadEnvironment", func() {
			master := "zk://local.host:2182/mesos"
			dbPath := "db/eremetic.db"
			frameworkID := "a_framework_id"
			httpCredentials := "admin:admin"
			urlPrefix := "/service/eremetic"

			os.Setenv("MASTER", master)
			os.Setenv("DATABASE", dbPath)
			os.Setenv("FRAMEWORK_ID", frameworkID)
			os.Setenv("HTTP_CREDENTIALS", httpCredentials)
			os.Setenv("URL_PREFIX", urlPrefix)

			ReadEnvironment(conf)

			So(conf.Master, ShouldEqual, master)
			So(conf.DatabasePath, ShouldEqual, dbPath)
			So(conf.FrameworkID, ShouldEqual, frameworkID)
			So(conf.HTTPCredentials, ShouldEqual, httpCredentials)
			So(conf.URLPrefix, ShouldEqual, urlPrefix)
		})
	})
}
