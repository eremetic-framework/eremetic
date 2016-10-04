package mesos

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func mockStatusWithSandbox() []byte {
	return []byte(`[
                    {
                      "Mounts": [
                        {
                          "Source": "/tmp/mesos/slaves/<agent_id>/frameworks/<framework_id>/executors/<task_id>/runs/<container_id>",
                          "Destination": "/mnt/mesos/sandbox",
                          "Mode": "",
                          "RW": true
                        }
                      ]
                    }
                  ]`)
}

func mockStatusWithoutSandbox() []byte {
	return []byte(`[
                    {
                      "Mounts": [
                        {
                          "Source": "/tmp/mesos/",
                          "Destination": "/mnt/not/the/sandbox",
                          "Mode": "",
                          "RW": true
                        }
                      ]
                    }
                  ]`)
}

func mockStatusNoMounts() []byte {
	return []byte(`[
                    {
                      "Mounts": []
                    }
                  ]`)
}

func TestExtractor(t *testing.T) {
	Convey("extractSandboxPath", t, func() {
		Convey("Sandbox found", func() {
			status := mockStatusWithSandbox()
			sandbox, err := extractSandboxPath(status)
			So(err, ShouldBeNil)
			So(sandbox, ShouldNotBeEmpty)
		})

		Convey("Sandbox not found", func() {
			status := mockStatusWithoutSandbox()
			sandbox, err := extractSandboxPath(status)
			So(sandbox, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("No mounts in data", func() {
			status := mockStatusWithoutSandbox()
			sandbox, err := extractSandboxPath(status)
			So(sandbox, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("Empty data", func() {
			sandbox, err := extractSandboxPath([]byte(""))
			So(sandbox, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}
