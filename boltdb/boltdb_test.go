package boltdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/klarna/eremetic"
)

func TestBoltDatabase(t *testing.T) {
	var (
		testDB string
		db     *TaskDB
	)

	setup := func() error {
		dir, _ := ioutil.TempDir("", "eremetic")
		testDB = fmt.Sprintf("%s/test.db", dir)
		adb, err := newCustomTaskDB(defaultConnector{}, testDB)
		if err != nil {
			return err
		}

		db = adb

		return nil
	}

	teardown := func() {
		os.Remove(testDB)
	}

	status := []eremetic.Status{
		eremetic.Status{
			Status: eremetic.TaskRunning,
			Time:   time.Now().Unix(),
		},
	}

	Convey("NewDB", t, func() {
		Convey("With an absolute path", func() {
			setup()
			defer teardown()

			So(db.conn.Path(), ShouldNotBeEmpty)
			So(filepath.IsAbs(db.conn.Path()), ShouldBeTrue)
		})
	})

	Convey("createBoltDriver", t, func() {
		Convey("Error", func() {
			setup()
			defer teardown()
			defer db.Close()

			connector := new(mockConnector)
			_, err := newCustomTaskDB(connector, "")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "missing boltdb database location")
		})
	})

	Convey("Close", t, func() {
		setup()
		defer teardown()
		db.Close()

		So(db.conn.Path(), ShouldBeEmpty)
	})

	Convey("Clean", t, func() {
		setup()
		defer teardown()
		defer db.Close()

		db.PutTask(&eremetic.Task{ID: "1234"})
		task, _ := db.ReadTask("1234")
		So(task, ShouldNotEqual, eremetic.Task{})
		So(task.ID, ShouldNotBeEmpty)

		db.Clean()

		task, _ = db.ReadTask("1234")
		So(task, ShouldBeZeroValue)
	})

	Convey("Put and Read Task", t, func() {
		setup()
		defer teardown()
		defer db.Close()

		var maskedEnv = make(map[string]string)
		maskedEnv["foo"] = "bar"

		task1 := eremetic.Task{ID: "1234"}
		task2 := eremetic.Task{
			ID:                "12345",
			TaskCPUs:          2.5,
			TaskMem:           15.3,
			Name:              "request Name",
			Status:            status,
			FrameworkID:       "1234",
			Command:           "echo date",
			User:              "root",
			Image:             "busybox",
			MaskedEnvironment: maskedEnv,
		}

		db.PutTask(&task1)
		db.PutTask(&task2)

		t1, err := db.ReadTask(task1.ID)
		So(t1, ShouldResemble, task1)
		So(err, ShouldBeNil)
		t2, err := db.ReadTask(task2.ID)
		So(err, ShouldBeNil)
		So(t2.MaskedEnvironment["foo"], ShouldEqual, "*******")
	})

	Convey("Read unmasked task", t, func() {
		setup()
		defer teardown()
		defer db.Close()

		var maskedEnv = make(map[string]string)
		maskedEnv["foo"] = "bar"

		task := eremetic.Task{
			ID:                "12345",
			TaskCPUs:          2.5,
			TaskMem:           15.3,
			Name:              "request Name",
			Status:            status,
			FrameworkID:       "1234",
			Command:           "echo date",
			User:              "root",
			Image:             "busybox",
			MaskedEnvironment: maskedEnv,
		}
		db.PutTask(&task)

		t, err := db.ReadUnmaskedTask(task.ID)
		So(t, ShouldResemble, task)
		So(err, ShouldBeNil)
		So(t.MaskedEnvironment, ShouldContainKey, "foo")
		So(t.MaskedEnvironment["foo"], ShouldEqual, "bar")

	})

	Convey("List non-terminal tasks", t, func() {
		setup()
		defer teardown()
		defer db.Close()

		db.Clean()

		// A terminated task
		db.PutTask(&eremetic.Task{
			ID: "1234",
			Status: []eremetic.Status{
				eremetic.Status{
					Status: eremetic.TaskStaging,
					Time:   time.Now().Unix(),
				},
				eremetic.Status{
					Status: eremetic.TaskRunning,
					Time:   time.Now().Unix(),
				},
				eremetic.Status{
					Status: eremetic.TaskFinished,
					Time:   time.Now().Unix(),
				},
			},
		})

		// A running task
		db.PutTask(&eremetic.Task{
			ID: "2345",
			Status: []eremetic.Status{
				eremetic.Status{
					Status: eremetic.TaskStaging,
					Time:   time.Now().Unix(),
				},
				eremetic.Status{
					Status: eremetic.TaskRunning,
					Time:   time.Now().Unix(),
				},
			},
		})

		tasks, err := db.ListNonTerminalTasks()
		So(err, ShouldBeNil)
		So(tasks, ShouldHaveLength, 1)
		task := tasks[0]
		So(task.ID, ShouldEqual, "2345")
	})

	Convey("DeleteTask", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()
			defer db.Close()
			db.Clean()

			var maskedEnv = make(map[string]string)
			maskedEnv["foo"] = "bar"

			task1 := eremetic.Task{ID: "1234"}
			db.PutTask(&task1)
			t1, err := db.ReadUnmaskedTask(task1.ID)
			So(t1, ShouldResemble, task1)
			So(err, ShouldBeNil)

			err = db.DeleteTask(task1.ID)
			So(err, ShouldBeNil)
		})
	})

	Convey("List non-terminal tasks no running task", t, func() {
		setup()
		defer teardown()
		defer db.Close()

		db.Clean()
		db.PutTask(&eremetic.Task{
			ID: "1234",
			Status: []eremetic.Status{
				eremetic.Status{
					Status: eremetic.TaskStaging,
					Time:   time.Now().Unix(),
				},
				eremetic.Status{
					Status: eremetic.TaskRunning,
					Time:   time.Now().Unix(),
				},
				eremetic.Status{
					Status: eremetic.TaskFinished,
					Time:   time.Now().Unix(),
				},
			},
		})
		tasks, err := db.ListNonTerminalTasks()
		So(err, ShouldBeNil)
		So(tasks, ShouldBeEmpty)
		So(tasks, ShouldNotBeNil)
	})
}
