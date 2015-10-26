package routes

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := Create()
			for _, r := range routes {
				So(m.GetRoute(r.Name), ShouldNotBeNil)
			}
		})
	})
}
