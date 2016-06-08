package database

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/klarna/eremetic/mocks"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/samuel/go-zookeeper/zk"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestZKDatabase(t *testing.T) {
	var (
		db        *zkDriver
		object    *mocks.ZkConnection
		connector *mocks.ZkConnectorInterface
	)

	zkPath := "zk://localhost:1234/testdb"

	setup := func() {
		db = &zkDriver{
			connection: new(mocks.ZkConnection),
			path:       "/testdb",
		}
		object = db.connection.(*mocks.ZkConnection)
		connector = new(mocks.ZkConnectorInterface)
	}

	teardown := func() {
		db = nil
		object = nil
		connector = nil
	}

	status := []types.Status{
		types.Status{
			Status: mesos.TaskState_TASK_RUNNING.String(),
			Time:   time.Now().Unix(),
		},
	}

	var maskedEnv = make(map[string]string)
	maskedEnv["foo"] = "bar"

	task := &types.EremeticTask{
		ID:                "1234",
		MaskedEnvironment: maskedEnv,
		Status:            status,
	}

	taskBytes, err := encode(task)
	if err != nil {
		t.Fail()
	}

	Convey("Creating", t, func() {
		Convey("Errors", func() {
			Convey("Missing path", func() {
				setup()
				defer teardown()

				_, err := createZKDriver(connector, "")

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Missing ZK path")
			})

			Convey("Unable to connect", func() {
				setup()
				defer teardown()

				connector.On("Connect", mock.AnythingOfType("string")).Return(nil, errors.New("Unable to connect"))

				_, err := createZKDriver(connector, zkPath)

				So(err, ShouldNotBeNil)
				So(connector.AssertCalled(t, "Connect", "localhost:1234"), ShouldBeTrue)
			})

			Convey("Unable to verify existance", func() {
				setup()
				defer teardown()
				connector.On("Connect", mock.AnythingOfType("string")).Return(object, nil)
				object.On("Exists", mock.AnythingOfType("string")).Return(false, &zk.Stat{}, errors.New("Bad Connection"))

				_, err := createZKDriver(connector, zkPath)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Bad Connection")
				So(connector.AssertCalled(t, "Connect", "localhost:1234"), ShouldBeTrue)
				So(object.AssertCalled(t, "Exists", "/testdb"), ShouldBeTrue)
			})

			Convey("Fail to create if not exists", func() {
				setup()
				defer teardown()
				connector.On("Connect", mock.AnythingOfType("string")).Return(object, nil)
				object.On("Exists", mock.AnythingOfType("string")).Return(false, &zk.Stat{}, nil)
				object.On("Create", mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("int32"), mock.Anything).Return("", errors.New("Unable to create node"))

				_, err := createZKDriver(connector, zkPath)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Unable to create node")
				So(connector.AssertCalled(t, "Connect", "localhost:1234"), ShouldBeTrue)
				So(object.AssertCalled(t, "Exists", "/testdb"), ShouldBeTrue)
				So(object.AssertCalled(t, "Create", "/testdb", mock.Anything, mock.AnythingOfType("int32"), mock.Anything), ShouldBeTrue)
			})
		})

		Convey("Success", func() {
			setup()
			defer teardown()
			connector.On("Connect", mock.AnythingOfType("string")).Return(object, nil)
			object.On("Exists", mock.AnythingOfType("string")).Return(true, &zk.Stat{}, nil)

			db, err := createZKDriver(connector, zkPath)

			So(err, ShouldBeNil)
			So(connector.AssertCalled(t, "Connect", "localhost:1234"), ShouldBeTrue)
			So(object.AssertCalled(t, "Exists", "/testdb"), ShouldBeTrue)
			So(db, ShouldImplement, (*TaskDB)(nil)) // Most weirdest syntax ever?
		})
	})

	Convey("Clean", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()

			object.On("Delete", mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(nil)

			err := db.Clean()

			So(err, ShouldBeNil)
			So(object.AssertCalled(t, "Delete", "/testdb/", mock.Anything), ShouldBeTrue)
		})
	})

	Convey("Close", t, func() {
		setup()
		defer teardown()

		object.On("Close").Return(nil)

		db.Close()

		So(object.AssertCalled(t, "Close"), ShouldBeTrue)
	})

	Convey("AddTask", t, func() {
		Convey("Exists", func() {
			setup()
			defer teardown()

			object.On("Exists", mock.AnythingOfType("string")).Return(true, &zk.Stat{}, nil)
			object.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(&zk.Stat{}, nil)

			err := db.PutTask(task)

			So(err, ShouldBeNil)
			So(object.AssertCalled(t, "Exists", "/testdb/1234"), ShouldBeTrue)
			So(object.AssertCalled(t, "Set", "/testdb/1234", taskBytes, mock.AnythingOfType("int32")), ShouldBeTrue)
		})

		Convey("New", func() {
			setup()
			defer teardown()

			object.On("Exists", mock.AnythingOfType("string")).Return(false, &zk.Stat{}, nil)
			object.On("Create", mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("int32"), mock.Anything).Return("", nil)

			err := db.PutTask(task)

			So(err, ShouldBeNil)
			So(object.AssertCalled(t, "Exists", "/testdb/1234"), ShouldBeTrue)
			So(object.AssertCalled(t, "Create", "/testdb/1234", taskBytes, mock.AnythingOfType("int32"), mock.Anything), ShouldBeTrue)
		})

		Convey("Errors", func() {
			Convey("Bad Connection", func() {
				setup()
				defer teardown()

				object.On("Exists", mock.AnythingOfType("string")).Return(false, &zk.Stat{}, errors.New("Bad Connection"))

				err := db.PutTask(task)
				So(err.Error(), ShouldEqual, "Bad Connection")
			})
		})
	})

	Convey("ReadUnmaskedTask", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()

			object.On("Get", mock.AnythingOfType("string")).Return(taskBytes, &zk.Stat{}, nil)

			read, err := db.ReadUnmaskedTask("1234")

			So(err, ShouldBeNil)
			So(&read, ShouldResemble, task)
			So(read.MaskedEnvironment["foo"], ShouldEqual, "bar")
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
		})

		Convey("Error", func() {
			setup()
			defer teardown()

			object.On("Get", mock.AnythingOfType("string")).Return([]byte{}, &zk.Stat{}, errors.New("Unable to Read"))

			_, err := db.ReadUnmaskedTask("1234")
			So(err.Error(), ShouldEqual, "Unable to Read")
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
		})
	})

	Convey("ReadTask", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()

			object.On("Get", mock.AnythingOfType("string")).Return(taskBytes, &zk.Stat{}, nil)

			read, err := db.ReadTask("1234")

			So(err, ShouldBeNil)
			So(&read, ShouldHaveSameTypeAs, task)
			So(read.ID, ShouldEqual, task.ID)
			So(read.MaskedEnvironment["foo"], ShouldEqual, masking)
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
		})

		Convey("Error", func() {
			setup()
			defer teardown()

			object.On("Get", mock.AnythingOfType("string")).Return([]byte{}, &zk.Stat{}, errors.New("Unable to Read"))

			_, err := db.ReadTask("1234")
			So(err.Error(), ShouldEqual, "Unable to Read")
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
		})
	})

	Convey("ListNonTerminalTasks", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()

			object.On("Children", mock.AnythingOfType("string")).Return([]string{"1234"}, nil, nil)
			object.On("Get", mock.AnythingOfType("string")).Return(taskBytes, &zk.Stat{}, nil)

			list, err := db.ListNonTerminalTasks()

			So(err, ShouldBeNil)
			So(list, ShouldHaveLength, 1)
			So(list[0], ShouldHaveSameTypeAs, task)
			So(list[0].ID, ShouldEqual, task.ID)
			So(list[0].MaskedEnvironment["foo"], ShouldEqual, masking)
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
			So(object.AssertCalled(t, "Children", "/testdb"), ShouldBeTrue)
		})

		Convey("Error", func() {
			setup()
			defer teardown()

			object.On("Children", mock.AnythingOfType("string")).Return([]string{"1234"}, nil, nil)
			object.On("Get", mock.AnythingOfType("string")).Return([]byte{}, &zk.Stat{}, errors.New("Unable to Read"))

			list, _ := db.ListNonTerminalTasks()
			So(list, ShouldBeEmpty)
			So(object.AssertCalled(t, "Get", "/testdb/1234"), ShouldBeTrue)
			So(object.AssertCalled(t, "Children", "/testdb"), ShouldBeTrue)
		})
	})

	Convey("Count", t, func() {
		Convey("Success", func() {
			setup()
			defer teardown()

			object.On("Children", mock.AnythingOfType("string")).Return([]string{"1234"}, &zk.Stat{NumChildren: 1}, nil)

			count := db.Count()

			So(count, ShouldEqual, 1)
			So(object.AssertCalled(t, "Children", "/testdb"), ShouldBeTrue)
		})

		Convey("Error", func() {
			setup()
			defer teardown()

			object.On("Children", mock.AnythingOfType("string")).Return([]string{"1234"}, &zk.Stat{NumChildren: 0}, errors.New("Unable to Read"))

			count := db.Count()
			So(count, ShouldEqual, 0)
		})
	})

	Convey("parsePath", t, func() {
		masters := make(map[string]string)
		masters["master1.local:1111,master2.local:1111,master3.local:1111"] =
			"zk://master1.local:1111,master2.local:1111,master3.local:1111/mesos"
		masters["master1.local:2222,master2.local:2222,master3.local:2222"] =
			"zk://master1.local:2222,master2.local:2222,master3.local:2222/cluster/mesos"
		masters["master1.local:3333,master2.local:3333,master3.local:3333"] =
			"zk://master1.local:3333,master2.local:3333,master3.local:3333/mesos/cluster"
		masters["123.123.123.123:4444,10.1.15.4:4444,10.1.15.42:4444"] =
			"zk://123.123.123.123:4444,10.1.15.4:4444,10.1.15.42:4444/"

		for e, p := range masters {
			Convey(p, func() {
				servers, paths, err := parsePath(p)
				So(err, ShouldBeNil)
				So(servers, ShouldEqual, e)
				So(strings.HasSuffix(p, paths), ShouldBeTrue)
			})
		}

		Convey("Ensures it starts with a / but doesn't end with a /", func() {
			_, path, err := parsePath("zk://master1.local:1111/pathing/")
			So(err, ShouldBeNil)
			So(strings.HasPrefix(path, "/"), ShouldBeTrue)
			So(strings.HasSuffix(path, "/"), ShouldBeFalse)
		})
	})
}
