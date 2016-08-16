package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	Convey("WasRunning", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.WasRunning(), ShouldBeTrue)
		})

		Convey("A task that never was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.WasRunning(), ShouldBeFalse)
		})
	})

	Convey("IsTerminated", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsTerminated(), ShouldBeFalse)
		})

		Convey("A task that never was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_FAILED"},
				},
			}

			So(task.IsTerminated(), ShouldBeTrue)
		})

		Convey("A empty task", func() {
			task := EremeticTask{}

			So(task.IsTerminated(), ShouldBeTrue)
		})
	})

	Convey("IsRunning", t, func() {
		Convey("A task that was running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
					Status{2, "TASK_FINISHED"},
				},
			}

			So(task.IsRunning(), ShouldBeFalse)
		})

		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{0, "TASK_STAGING"},
					Status{1, "TASK_RUNNING"},
				},
			}

			So(task.IsRunning(), ShouldBeTrue)
		})

		Convey("A empty task", func() {
			task := EremeticTask{}

			So(task.IsRunning(), ShouldBeFalse)
		})
	})

	Convey("LastUpdated", t, func() {
		Convey("A task that is running", func() {
			task := EremeticTask{
				Status: []Status{
					Status{1449682262, "TASK_STAGING"},
					Status{1449682265, "TASK_RUNNING"},
				},
			}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 1449682265)
		})

		Convey("A empty task", func() {
			task := EremeticTask{}

			s := task.LastUpdated()

			So(s.Unix(), ShouldEqual, 0)
		})
	})

	Convey("NewEremeticTask", t, func() {
		request := Request{
			TaskCPUs:    0.5,
			TaskMem:     22.0,
			DockerImage: "busybox",
			Command:     "echo hello",
		}

		Convey("No volume or environment specified", func() {
			task, err := NewEremeticTask(request, "")

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
			var volumes []Volume
			var environment = make(map[string]string)
			environment["foo"] = "bar"
			volumes = append(volumes, Volume{
				ContainerPath: "/var/www",
				HostPath:      "/var/www",
			})
			request.Volumes = volumes
			request.Environment = environment

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.Environment, ShouldContainKey, "foo")
			So(task.Environment["foo"], ShouldEqual, "bar")
			So(task.Volumes[0].ContainerPath, ShouldEqual, volumes[0].ContainerPath)
			So(task.Volumes[0].HostPath, ShouldEqual, volumes[0].HostPath)
		})

		Convey("Given a masked environment", func() {
			var maskedEnv = make(map[string]string)
			maskedEnv["foo"] = "bar"

			request.MaskedEnvironment = maskedEnv
			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.MaskedEnvironment, ShouldContainKey, "foo")
			So(task.MaskedEnvironment["foo"], ShouldEqual, "bar")
		})

		Convey("Given URI (via uris) to download", func() {
			request.URIs = []string{"http://foobar.local/kitten.jpg"}

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.URIs[0])
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeFalse)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via uris) to download and extract", func() {
			request.URIs = []string{"http://foobar.local/kittens.tar.gz"}

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.URIs[0])
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeTrue)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via fetch) to download and extract", func() {
			request.Fetch = []URI{URI{
				URI:     "http://foobar.local/kittens.tar.gz",
				Extract: true,
			}}

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.Fetch[0].URI)
			So(task.FetchURIs[0].Executable, ShouldBeFalse)
			So(task.FetchURIs[0].Extract, ShouldBeTrue)
			So(task.FetchURIs[0].Cache, ShouldBeFalse)
		})

		Convey("Given URI (via fetch) to download, cache and set executable", func() {
			request.Fetch = []URI{URI{
				URI:        "http://foobar.local/kittens.sh",
				Executable: true,
				Cache:      true,
			}}

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.FetchURIs, ShouldHaveLength, 1)
			So(task.FetchURIs[0].URI, ShouldEqual, request.Fetch[0].URI)
			So(task.FetchURIs[0].Executable, ShouldBeTrue)
			So(task.FetchURIs[0].Extract, ShouldBeFalse)
			So(task.FetchURIs[0].Cache, ShouldBeTrue)
		})

		Convey("Given no Command", func() {
			request.Command = ""

			task, err := NewEremeticTask(request, "")

			So(err, ShouldBeNil)
			So(task.Command, ShouldBeEmpty)
		})

		Convey("New task from empty request", func() {
			req := Request{}
			task, err := NewEremeticTask(req, "")

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			So(task.WasRunning(), ShouldBeFalse)
			So(task.IsRunning(), ShouldBeFalse)
			So(task.IsTerminated(), ShouldBeFalse)
		})
	})
}
