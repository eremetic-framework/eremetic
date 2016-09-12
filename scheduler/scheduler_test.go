package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/database"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func callbackReceiver() (chan callbackData, *httptest.Server) {
	cb := make(chan callbackData, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var callback callbackData
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			close(cb)
			return
		}
		err = json.Unmarshal(body, &callback)
		if err != nil {
			close(cb)
			return
		}
		fmt.Fprintln(w, "ok")
		cb <- callback
	}))
	return cb, ts
}

func TestScheduler(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	dir, _ := os.Getwd()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil || db == nil {
		t.Error("Foo")
		t.Fail()
	}
	db.Clean()
	defer db.Close()

	Convey("Scheduling", t, func() {
		Convey("Given a scheduler with one scheduled task", func() {
			s := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			Convey("When creating a new scheduler", func() {
				s = Create(&Settings{MaxQueueSize: 200}, db)

				Convey("The settings should have default values", func() {
					So(s.tasksCreated, ShouldEqual, 0)
					So(cap(s.tasks), ShouldEqual, 200)
				})
			})

			Convey("When the scheduler is registered", func() {
				driver := NewMockScheduler()
				driver.On("ReconcileTasks").Return("ok").Once()

				fid := mesos.FrameworkID{Value: proto.String("1234")}
				info := mesos.MasterInfo{}

				s.Registered(driver, &fid, &info)

				Convey("The tasks should be reconciled", func() {
					So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
				})
			})

			Convey("When the scheduler is reregistered", func() {
				driver := NewMockScheduler()
				driver.On("ReconcileTasks").Return("ok").Once()
				db.Clean()

				s.Reregistered(driver, &mesos.MasterInfo{})

				Convey("The tasks should be reconciled", func() {
					So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
				})
			})

			Convey("When the scheduler is disconnected", func() {
				s.Disconnected(nil)
			})

			Convey("When an offer rescinded", func() {
				s.OfferRescinded(nil, &mesos.OfferID{})
			})

			Convey("When a slave was lost", func() {
				s.SlaveLost(nil, &mesos.SlaveID{})
			})

			Convey("When an executor was lost", func() {
				s.ExecutorLost(nil, &mesos.ExecutorID{}, &mesos.SlaveID{}, 2)
			})

			Convey("When there was an error", func() {
				s.Error(nil, "Error")
			})
		})
	})
	Convey("ResourceOffers", t, func() {
		Convey("Given a scheduler with one scheduled task", func() {
			s := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			driver := NewMockScheduler()

			Convey("When there are no offers", func() {
				offers := []*mesos.Offer{}
				s.ResourceOffers(driver, offers)

				Convey("No offers should be declined", func() {
					So(driver.AssertNotCalled(t, "DeclineOffer"), ShouldBeTrue)
				})
				Convey("No tasks should be launched", func() {
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})
			})
			Convey("When there are no tasks", func() {
				offers := []*mesos.Offer{
					offer("1234", 1.0, 128),
				}
				driver.On("DeclineOffer").Return("declined").Once()

				s.ResourceOffers(driver, offers)

				Convey("The offer should be declined", func() {
					So(driver.AssertCalled(t, "DeclineOffer"), ShouldBeTrue)
				})
				Convey("No tasks should be launched", func() {
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})
			})
			Convey("When a task is able to launch", func() {
				offers := []*mesos.Offer{
					offer("1234", 1.0, 128),
				}
				driver.On("LaunchTasks").Return("launched").Once()

				taskID, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(driver, offers)

				task, err := db.ReadTask(taskID)
				So(err, ShouldBeNil)

				Convey("The task should contain the status history", func() {
					So(task.Status, ShouldHaveLength, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskState_TASK_QUEUED)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskState_TASK_STAGING)
				})
				Convey("The offer should not be declined", func() {
					So(driver.AssertNotCalled(t, "DeclineOffer"), ShouldBeTrue)
				})
				Convey("The tasks should be launched", func() {
					So(driver.AssertCalled(t, "LaunchTasks"), ShouldBeTrue)
				})
			})

			Convey("When a task unable to launch", func() {
				offers := []*mesos.Offer{
					offer("1234", 1.0, 128),
				}
				driver.On("DeclineOffer").Return("declined").Once()

				_, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    1.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(driver, offers)

				Convey("The offer should be declined", func() {
					So(driver.AssertCalled(t, "DeclineOffer"), ShouldBeTrue)
				})
				Convey("The tasks should not be launched", func() {
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})
			})
		})
	})
	Convey("StatusUpdate", t, func() {
		Convey("Given a scheduler with one scheduled task", func() {
			s := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			Convey("When a running task fails", func() {
				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_RUNNING.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				So(len(task.Status), ShouldEqual, 1)
				So(task.Status[0].Status, ShouldEqual, eremetic.TaskState_TASK_RUNNING)

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FAILED.Enum(),
				})

				task, err = db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("The task status history should contain the failed status", func() {
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskState_TASK_RUNNING)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskState_TASK_FAILED)
				})
			})

			Convey("When a task fails immediately", func() {
				s.tasks = make(chan string, 100)

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FAILED.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("The task should have a queued status", func() {
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskState_TASK_FAILED)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskState_TASK_QUEUED)
				})

				Convey("The task should be published on channel", func() {
					c := <-s.tasks

					So(c, ShouldEqual, id)
				})
			})

			Convey("When a task finishes with a callback specified", func() {
				cb, ts := callbackReceiver()
				defer ts.Close()

				id := "eremetic-task.1000"

				db.PutTask(&eremetic.Task{
					ID:          id,
					CallbackURI: ts.URL,
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FINISHED.Enum(),
				})

				Convey("The callback data should be available", func() {
					c := <-cb

					So(c.TaskID, ShouldEqual, id)
					So(c.Status, ShouldEqual, "TASK_FINISHED")
				})
			})

			Convey("When a task fails with a callback specified", func() {
				cb, ts := callbackReceiver()
				defer ts.Close()

				id := "eremetic-task.1001"

				db.PutTask(&eremetic.Task{
					ID:          id,
					CallbackURI: ts.URL,
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FAILED.Enum(),
				})

				Convey("The callback data should be available", func() {
					c := <-cb

					So(c.TaskID, ShouldEqual, id)
					So(c.Status, ShouldEqual, "TASK_FAILED")
				})
			})

			Convey("When a task retries with a callback specified", func() {
				cb, ts := callbackReceiver()
				defer ts.Close()

				id := "eremetic-task.1002"

				db.PutTask(&eremetic.Task{
					ID:          id,
					CallbackURI: ts.URL,
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FAILED.Enum(),
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					State: mesos.TaskState_TASK_FINISHED.Enum(),
				})

				Convey("The callback data should be available", func() {
					c := <-cb

					So(c.TaskID, ShouldEqual, id)
					So(c.Status, ShouldEqual, "TASK_FINISHED")
				})
			})

			Convey("When the sandbox is updated", func() {
				id := "eremetic-task.1003"

				s.StatusUpdate(nil, &mesos.TaskStatus{
					TaskId: &mesos.TaskID{
						Value: proto.String(id),
					},
					Data:  []byte(`[{"Mounts":[{"Source":"/tmp/mesos/slaves/<agent_id>/frameworks/<framework_id>/executors/<task_id>/runs/<container_id>","Destination":"/mnt/mesos/sandbox","Mode":"","RW":true}]}]`),
					State: mesos.TaskState_TASK_RUNNING.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("There should be a path to the sandbox", func() {
					So(task.SandboxPath, ShouldNotBeEmpty)
				})
			})
		})
	})
	Convey("FrameworkMessage", t, func() {
		Convey("Given a scheduler with one scheduled task", func() {
			s := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"
			db.PutTask(&eremetic.Task{ID: id})

			driver := NewMockScheduler()
			message := `{"message": "this is a message"}`

			Convey("When receiving a framework message from Eremetic", func() {
				source := "eremetic-executor"
				executor := mesos.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, message)
			})

			Convey("When receiving a framework message from an unknown source", func() {
				source := "other-source"
				executor := mesos.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, message)
			})

			Convey("When the framework message format is invalid", func() {
				source := "eremetic-executor"
				executor := mesos.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, "not a json")
			})
		})
	})
	Convey("ScheduleTask", t, func() {
		Convey("Given a scheduler with no scheduled tasks", func() {
			scheduler := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			Convey("When scheduling a task", func() {
				request := eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				}

				taskID, err := scheduler.ScheduleTask(request)
				So(err, ShouldBeNil)

				Convey("It should put a task id on the channel", func() {
					c := <-scheduler.tasks
					So(c, ShouldEqual, taskID)
				})

				Convey("The task should be present in the database", func() {
					task, err := db.ReadTask(taskID)
					So(err, ShouldBeNil)

					So(task.TaskCPUs, ShouldEqual, request.TaskCPUs)
					So(task.TaskMem, ShouldEqual, request.TaskMem)
					So(task.Command, ShouldEqual, request.Command)
					So(task.User, ShouldEqual, "root")
					So(task.Environment, ShouldBeEmpty)
					So(task.Image, ShouldEqual, request.DockerImage)
					So(task.ID, ShouldStartWith, "eremetic-task.")
				})
			})

			Convey("When scheduling a task and the queue is full", func() {
				scheduler.tasks <- "dummy"

				request := eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				}

				_, err := scheduler.ScheduleTask(request)

				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
	Convey("nextID", t, func() {
		Convey("Given a scheduler with no scheduled tasks", func() {
			scheduler := &eremeticScheduler{
				tasks:    make(chan string, 100),
				database: db,
			}

			Convey("When generating a random string", func() {
				str := nextID(scheduler)

				Convey("The string should not be empty", func() {
					So(str, ShouldNotBeEmpty)
				})
			})
		})
	})
}

//------------------ Mock Scheduler ------------------------------------------//

type MockScheduler struct {
	mock.Mock
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{}
}

func (sched *MockScheduler) Abort() (stat mesos.Status, err error) {
	sched.Called()
	return mesos.Status_DRIVER_ABORTED, nil
}

func (sched *MockScheduler) AcceptOffers(offerIds []*mesos.OfferID, operations []*mesos.Offer_Operation, filters *mesos.Filters) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) DeclineOffer(*mesos.OfferID, *mesos.Filters) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_STOPPED, nil
}

func (sched *MockScheduler) Join() (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) KillTask(*mesos.TaskID) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) ReconcileTasks([]*mesos.TaskStatus) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) RequestResources([]*mesos.Request) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) ReviveOffers() (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) Run() (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) Start() (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) Stop(bool) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) SendFrameworkMessage(*mesos.ExecutorID, *mesos.SlaveID, string) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) LaunchTasks([]*mesos.OfferID, []*mesos.TaskInfo, *mesos.Filters) (mesos.Status, error) {
	sched.Called()
	return mesos.Status_DRIVER_RUNNING, nil
}

func (sched *MockScheduler) Registered(sched.SchedulerDriver, *mesos.FrameworkID, *mesos.MasterInfo) {
	sched.Called()
}

func (sched *MockScheduler) Reregistered(sched.SchedulerDriver, *mesos.MasterInfo) {
	sched.Called()
}

func (sched *MockScheduler) Disconnected(sched.SchedulerDriver) {
	sched.Called()
}

func (sched *MockScheduler) ResourceOffers(sched.SchedulerDriver, []*mesos.Offer) {
	sched.Called()
}

func (sched *MockScheduler) OfferRescinded(sched.SchedulerDriver, *mesos.OfferID) {
	sched.Called()
}

func (sched *MockScheduler) StatusUpdate(sched.SchedulerDriver, *mesos.TaskStatus) {
	sched.Called()
}

func (sched *MockScheduler) FrameworkMessage(sched.SchedulerDriver, *mesos.ExecutorID, *mesos.SlaveID, string) {
	sched.Called()
}

func (sched *MockScheduler) SlaveLost(sched.SchedulerDriver, *mesos.SlaveID) {
	sched.Called()
}

func (sched *MockScheduler) ExecutorLost(sched.SchedulerDriver, *mesos.ExecutorID, *mesos.SlaveID, int) {
	sched.Called()
}

func (sched *MockScheduler) Error(d sched.SchedulerDriver, _msg string) {
	sched.Called()
}
