package routes

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	routes := []string{"AddTask", "Status"}

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := Create(nil)
			for _, name := range routes {
				So(m.GetRoute(name), ShouldNotBeNil)
			}
		})
	})
}
