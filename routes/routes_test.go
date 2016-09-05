package routes

import (
	"fmt"
	"os"
	"testing"

	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	routes := routes(handler.Handler{}, &config.Config{})
	dir := os.TempDir()
	db, _ := database.NewDB("boltdb", fmt.Sprintf("%s/eremetic_test.db", dir))

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := Create(nil, &config.Config{}, db)
			for _, route := range routes {
				So(m.GetRoute(route.Name), ShouldNotBeNil)
			}
		})
	})

	Convey("Expected number of routes", t, func() {
		ExpectedNumberOfRoutes := 8 // Magic numbers FTW

		So(len(routes), ShouldEqual, ExpectedNumberOfRoutes)
	})
}
