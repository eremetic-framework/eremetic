package mesos

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic"
	"github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {

	status := []eremetic.Status{
		eremetic.Status{
			Status: eremetic.TaskState_TASK_RUNNING,
			Time:   time.Now().Unix(),
		},
	}

	Convey("createTaskInfo", t, func() {
		eremeticTask := eremetic.Task{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "echo hello",
			Image:    "busybox",
			Status:   status,
			ID:       "eremetic-task.1234",
			Name:     "Eremetic task 17",
		}

		portres := "ports"
		offer := mesosproto.Offer{
			FrameworkId: &mesosproto.FrameworkID{
				Value: proto.String("framework-id"),
			},
			SlaveId: &mesosproto.SlaveID{
				Value: proto.String("slave-id"),
			},
			Hostname: proto.String("hostname"),
			Resources: []*mesosproto.Resource{&mesosproto.Resource{
				Name: &portres,
				Type: mesosproto.Value_RANGES.Enum(),
				Ranges: &mesosproto.Value_Ranges{
					Range: []*mesosproto.Value_Range{
						mesosutil.NewValueRange(31000, 31010),
					},
				},
			}},
		}

		Convey("No volume or environment specified", func() {
			net, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.GetName(), ShouldEqual, eremeticTask.Name)
			So(taskInfo.GetResources()[0].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskCPUs)
			So(taskInfo.GetResources()[1].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskMem)
			So(taskInfo.Container.GetType().String(), ShouldEqual, "DOCKER")
			So(taskInfo.Container.Docker.GetImage(), ShouldEqual, "busybox")
			So(net.SlaveId, ShouldEqual, "slave-id")
			So(taskInfo.Container.Docker.GetForcePullImage(), ShouldBeFalse)
		})

		Convey("Given no Command", func() {
			eremeticTask.Command = ""

			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.Command.GetValue(), ShouldBeEmpty)
			So(taskInfo.Command.GetShell(), ShouldBeFalse)
		})

		Convey("Given a volume and environment", func() {
			volumes := []eremetic.Volume{{
				ContainerPath: "/var/www",
				HostPath:      "/var/www",
			}}

			environment := make(map[string]string)
			environment["foo"] = "bar"

			eremeticTask.Environment = environment
			eremeticTask.Volumes = volumes

			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Container.Volumes[0].GetContainerPath(), ShouldEqual, volumes[0].ContainerPath)
			So(taskInfo.Container.Volumes[0].GetHostPath(), ShouldEqual, volumes[0].HostPath)
			So(taskInfo.Command.Environment.Variables[0].GetName(), ShouldEqual, "foo")
			So(taskInfo.Command.Environment.Variables[0].GetValue(), ShouldEqual, "bar")
			So(taskInfo.Command.Environment.Variables[1].GetName(), ShouldEqual, "MESOS_TASK_ID")
			So(taskInfo.Command.Environment.Variables[1].GetValue(), ShouldEqual, eremeticTask.ID)
		})

		Convey("Given a portMapping", func() {
			var ports []eremetic.Port

			ports = append(ports,
				eremetic.Port{
					ContainerPort: 80,
					Protocol:      "tcp",
				},
			)

			eremeticTask.Ports = ports

			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(len(taskInfo.Container.Docker.PortMappings), ShouldEqual, 1)
			So(taskInfo.Container.Docker.GetPortMappings()[0].GetContainerPort(), ShouldEqual, ports[0].ContainerPort)
			So(taskInfo.GetResources()[2].GetName(), ShouldEqual, "ports")

			expected_range := mesosutil.NewValueRange(31000, 31001)
			So(taskInfo.GetResources()[2].GetRanges().GetRange()[0].GetBegin(), ShouldEqual, expected_range.GetBegin())
			So(taskInfo.GetResources()[2].GetRanges().GetRange()[0].GetEnd(), ShouldEqual, expected_range.GetEnd())
		})

		Convey("Given archive to fetch", func() {
			URI := []eremetic.URI{{
				URI:     "http://foobar.local/cats.zip",
				Extract: true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Given archive to fetch and cache", func() {
			URI := []eremetic.URI{{
				URI:     "http://foobar.local/cats.zip",
				Extract: true,
				Cache:   true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeTrue)
		})

		Convey("Given image to fetch", func() {
			URI := []eremetic.URI{{
				URI: "http://foobar.local/cats.jpeg",
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Given script to fetch", func() {
			URI := []eremetic.URI{{
				URI:        "http://foobar.local/cats.sh",
				Executable: true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Force pull of docker image", func() {
			eremeticTask.ForcePullImage = true
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Container.Docker.GetForcePullImage(), ShouldBeTrue)
		})
	})
}
