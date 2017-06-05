package mesos

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDriver(t *testing.T) {
	Convey("createDriver", t, func() {
		Convey("Error when master URL can't be found", func() {
			scheduler := Scheduler{}

			driver, err := createDriver(&scheduler, &Settings{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Missing master location URL.")
			So(driver, ShouldBeNil)
		})
	})

	Convey("getFrameworkID", t, func() {
		Convey("Empty ID", func() {
			fid := getFrameworkID(&Scheduler{})
			So(fid, ShouldBeNil)
		})

		Convey("Some random string", func() {
			fid := getFrameworkID(&Scheduler{
				frameworkID: "zoidberg",
			})
			So(fid.Value, ShouldEqual, "zoidberg")
		})
	})
}
