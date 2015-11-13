package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

func setup() error {
	dir, _ := os.Getwd()
	return NewDB(fmt.Sprintf("%s/../db/test.db", dir))
}

func TestDatabase(t *testing.T) {
	Convey("NewDB", t, func() {
		Convey("With an absolute path", func() {
			err := setup()
			defer Close()

			So(boltdb.Path(), ShouldNotBeEmpty)
			So(filepath.IsAbs(boltdb.Path()), ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("With a relative path", func() {
			NewDB("db/test.db")
			defer Close()

			dir, _ := os.Getwd()
			So(filepath.IsAbs(boltdb.Path()), ShouldBeTrue)
			So(boltdb.Path(), ShouldEqual, fmt.Sprintf("%s/../db/test.db", dir))
		})
	})

	Convey("Close", t, func() {
		setup()
		Close()

		So(boltdb.Path(), ShouldBeEmpty)
	})

	Convey("Clean", t, func() {
		setup()
		defer Close()

		PutTask(&types.EremeticTask{ID: "1234"})
		task, _ := ReadTask("1234")
		So(task, ShouldNotEqual, types.EremeticTask{})
		So(task.ID, ShouldNotBeEmpty)

		Clean()

		task, _ = ReadTask("1234")
		So(task, ShouldBeZeroValue)
	})

	Convey("Put and Read Task", t, func() {
		setup()
		defer Close()

		task1 := types.EremeticTask{ID: "1234"}
		task2 := types.EremeticTask{
			ID:          "12345",
			TaskCPUs:    2.5,
			TaskMem:     15.3,
			Name:        "request Name",
			Status:      mesos.TaskState_TASK_STAGING.String(),
			FrameworkId: "1234",
			Command: &mesos.CommandInfo{
				Value: proto.String("echo date"),
				User:  proto.String("root"),
				Environment: &mesos.Environment{
					Variables: nil,
				},
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_DOCKER.Enum(),
				Docker: &mesos.ContainerInfo_DockerInfo{
					Image: proto.String("busybox"),
				},
				Volumes: nil,
			},
		}

		PutTask(&task1)
		PutTask(&task2)

		t1, err := ReadTask(task1.ID)
		So(t1, ShouldResemble, task1)
		So(err, ShouldBeNil)
		t2, err := ReadTask(task2.ID)
		So(t2, ShouldResemble, task2)
		So(err, ShouldBeNil)
	})
}
