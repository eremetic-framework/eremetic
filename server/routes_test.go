package server

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/config"
)

func TestRoutes(t *testing.T) {
	routes := routes(Handler{}, &config.Config{})

	db := eremetic.NewDefaultTaskDB()

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := NewRouter(nil, &config.Config{}, db)
			for _, route := range routes {
				So(m.GetRoute(route.Name), ShouldNotBeNil)
			}
		})
	})

	Convey("Expected number of routes", t, func() {
		ExpectedNumberOfRoutes := 18 // Magic numbers FTW

		So(len(routes), ShouldEqual, ExpectedNumberOfRoutes)
	})
}
