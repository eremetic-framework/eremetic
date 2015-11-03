package handler

import (
	"testing"
	"time"

	"github.com/alde/eremetic/types"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("createID", t, func() {
		Convey("Given a string", func() {
			Convey("It should build the appropriate ID", func() {
				So(createID("1234"), ShouldEqual, "eremetic-task.1234")
			})
		})
	})

	Convey("createEremeticTask", t, func() {
		request := types.Request{
			TaskCPUs:    0.5,
			TaskMem:     22.0,
			DockerImage: "busybox",
			Command:     "echo hello",
		}

		Convey("No volume or environment specified", func() {
			task, err := createEremeticTask(request)

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			So(task.Command.GetValue(), ShouldEqual, "echo hello")
			So(task.deleteAt, ShouldBeZeroValue)
			So(task.Container.GetType().String(), ShouldEqual, "DOCKER")
			So(task.Container.Docker.GetImage(), ShouldEqual, "busybox")
			So(task.Command.Environment.GetVariables(), ShouldBeEmpty)
			So(task.Container.Volumes, ShouldBeEmpty)
			So(task.Status, ShouldEqual, "TASK_STAGING")
		})

		Convey("Given a volume and environment", func() {
			var volumes []types.Volume
			var environment = make(map[string]string)
			environment["foo"] = "bar"
			volumes = append(volumes, types.Volume{
				ContainerPath: "/var/www",
				HostPath:      "/var/www",
			})
			request.Volumes = volumes
			request.Environment = environment

			task, err := createEremeticTask(request)

			So(err, ShouldBeNil)
			So(task.Command.Environment.Variables[0].GetName(), ShouldEqual, "foo")
			So(task.Command.Environment.Variables[0].GetValue(), ShouldEqual, environment["foo"])
			So(task.Container.Volumes[0].GetContainerPath(), ShouldEqual, volumes[0].ContainerPath)
			So(task.Container.Volumes[0].GetHostPath(), ShouldEqual, volumes[0].HostPath)
		})
	})

	Convey("createTaskInfo", t, func() {
		eremeticTask := eremeticTask{
			TaskCPUs:  0.2,
			TaskMem:   0.5,
			Command:   &mesos.CommandInfo{},
			Container: &mesos.ContainerInfo{},
			Status:    "TASK_RUNNING",
			ID:        "eremetic-task.1234",
			Name:      "Eremetic task 17",
			deleteAt:  time.Now(),
		}
		offer := mesos.Offer{
			FrameworkId: &mesos.FrameworkID{
				Value: proto.String("framework-id"),
			},
			SlaveId: &mesos.SlaveID{
				Value: proto.String("slave-id"),
			},
			Hostname: proto.String("hostname"),
		}

		taskInfo := createTaskInfo(&eremeticTask, &offer)
		So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
		So(taskInfo.GetName(), ShouldEqual, eremeticTask.Name)
		So(taskInfo.GetResources()[0].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskCPUs)
		So(taskInfo.GetResources()[1].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskMem)
	})
}
