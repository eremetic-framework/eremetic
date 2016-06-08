package routes

import (
	"fmt"
	"os"
	"testing"

	"github.com/klarna/eremetic/database"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	routes := []string{"AddTask", "Status"}

	dir, _ := os.Getwd()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil {
		t.Fail()
	}

	Convey("Create", t, func() {
		Convey("Should build the expected routes", func() {
			m := Create(nil, db)
			for _, name := range routes {
				So(m.GetRoute(name), ShouldNotBeNil)
			}
		})
	})
}
