package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDriver(t *testing.T) {
	Convey("createDriver", t, func() {
		Convey("Error when master URL can't be found", func() {
			scheduler := eremeticScheduler{}

			driver, err := createDriver(&scheduler)

			So(err.Error(), ShouldEqual, "Missing master location URL.")
			So(driver, ShouldBeNil)
		})
	})
}
