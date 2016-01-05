package scheduler

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {

	status := []types.Status{
		types.Status{
			Status: mesos.TaskState_TASK_RUNNING.String(),
			Time:   time.Now().Unix(),
		},
	}

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
			So(task.Command, ShouldEqual, "echo hello")
			So(task.User, ShouldEqual, "root")
			So(task.Environment, ShouldBeEmpty)
			So(task.Image, ShouldEqual, "busybox")
			So(task.Volumes, ShouldBeEmpty)
			So(task.Status[0].Status, ShouldEqual, "TASK_STAGING")
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
			So(task.Environment, ShouldContainKey, "foo")
			So(task.Environment["foo"], ShouldEqual, "bar")
			So(task.Volumes[0].ContainerPath, ShouldEqual, volumes[0].ContainerPath)
			So(task.Volumes[0].HostPath, ShouldEqual, volumes[0].HostPath)
		})
	})

	Convey("createTaskInfo", t, func() {
		volumes := []types.Volume{types.Volume{
			ContainerPath: "/var/www",
			HostPath:      "/var/www",
		}}

		environment := make(map[string]string)
		environment["foo"] = "bar"

		eremeticTask := types.EremeticTask{
			TaskCPUs:    0.2,
			TaskMem:     0.5,
			Command:     "echo hello",
			Environment: environment,
			Image:       "busybox",
			Volumes:     volumes,
			Status:      status,
			ID:          "eremetic-task.1234",
			Name:        "Eremetic task 17",
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

		net, taskInfo := createTaskInfo(eremeticTask, &offer)
		So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
		So(taskInfo.GetName(), ShouldEqual, eremeticTask.Name)
		So(taskInfo.GetResources()[0].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskCPUs)
		So(taskInfo.GetResources()[1].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskMem)
		So(taskInfo.Container.GetType().String(), ShouldEqual, "DOCKER")
		So(taskInfo.Container.Docker.GetImage(), ShouldEqual, "busybox")
		So(taskInfo.Container.Volumes[0].GetContainerPath(), ShouldEqual, volumes[0].ContainerPath)
		So(taskInfo.Container.Volumes[0].GetHostPath(), ShouldEqual, volumes[0].HostPath)
		So(taskInfo.Command.Environment.Variables[0].GetName(), ShouldEqual, "foo")
		So(taskInfo.Command.Environment.Variables[0].GetValue(), ShouldEqual, "bar")
		So(taskInfo.Command.Environment.Variables[1].GetName(), ShouldEqual, "MESOS_TASK_ID")
		So(taskInfo.Command.Environment.Variables[1].GetValue(), ShouldEqual, eremeticTask.ID)
		So(net.SlaveId, ShouldEqual, "slave-id")
	})
}
