package handler

import (
	"testing"
	"time"

	"github.com/alde/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("createEremeticTask", t, func() {
		request := types.Request{
			TaskCPUs:    0.5,
			TaskMem:     22.0,
			DockerImage: "busybox",
			Command:     "echo hello",
		}

		Convey("No volume or environment specified", func() {
			task := createEremeticTask(request)

			So(task, ShouldNotBeNil)
			So(task.Command.GetValue(), ShouldEqual, "echo hello")
			So(task.deleteAt, ShouldBeZeroValue)
			So(task.Container.GetType().String(), ShouldEqual, "DOCKER")
			So(task.Container.Docker.GetImage(), ShouldEqual, "busybox")
			So(task.Command.Environment.GetVariables(), ShouldBeEmpty)
			So(task.Container.Volumes, ShouldBeEmpty)
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

			task := createEremeticTask(request)

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
			deleteAt:  time.Now(),
		}
		offer := mesos.Offer{}

		taskInfo := createTaskInfo(&eremeticTask, 0, &offer)
		So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
		So(taskInfo.GetName(), ShouldEqual, "Eremetic task 0")
		So(taskInfo.GetResources()[0].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskCPUs)
		So(taskInfo.GetResources()[1].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskMem)
	})
}
