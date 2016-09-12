package server

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/klarna/eremetic/boltdb"
	"github.com/klarna/eremetic/config"
)

func TestRoutes(t *testing.T) {
	routes := routes(Handler{}, &config.Config{})
	dir := os.TempDir()

	db, err := boltdb.NewTaskDB(fmt.Sprintf("%s/eremetic_test.db", dir))
	if err != nil {
		t.Fatal(err)
	}

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := NewRouter(nil, &config.Config{}, db)
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
