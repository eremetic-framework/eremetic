package server

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatter(t *testing.T) {
	Convey("FormatTime", t, func() {
		Convey("A Valid Unix Timestamp", func() {
			t := time.Now().Unix()
			So(FormatTime(t), ShouldNotBeEmpty)
		})
	})
}
