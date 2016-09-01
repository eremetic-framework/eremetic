package routes

import (
	"testing"

	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	routes := routes(handler.Handler{}, &config.Config{})

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := Create(nil, &config.Config{})
			for _, r := range routes {
				So(m.GetRoute(r.Name), ShouldNotBeNil)
			}
		})
	})

	Convey("Expected number of routes", t, func() {
		ExpectedNumberOfRoutes := 8 // Magic numbers FTW

		So(len(routes), ShouldEqual, ExpectedNumberOfRoutes)
	})
}
