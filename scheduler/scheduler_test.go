package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic/database"
	"github.com/klarna/eremetic/types"
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
	dir, _ := os.Getwd()
	db, err := database.NewDB("boltdb", fmt.Sprintf("%s/../db/test.db", dir))
	if err != nil || db == nil {
		t.Error("Foo")
		t.Fail()
	}
	db.Clean()
	defer db.Close()

	Convey("eremeticScheduler", t, func() {
		s := &eremeticScheduler{
			tasks:    make(chan string, 1),
			database: db,
		}
		id := "eremetic-task.9999"
		db.PutTask(&types.EremeticTask{ID: id})

		Convey("newTask", func() {
			task := types.EremeticTask{
				ID: "eremetic-task.1234",
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
			taskData, mesosTask := s.newTask(task, &offer)

			So(mesosTask.GetTaskId().GetValue(), ShouldEqual, task.ID)
			So(taskData.SlaveId, ShouldEqual, "slave-id")
		})

		Convey("Create", func() {
			s := Create(&Settings{
				MaxQueueSize: 200,
			}, db)
			So(s.tasksCreated, ShouldEqual, 0)
			So(cap(s.tasks), ShouldEqual, 200)
		})

		Convey("API", func() {
			Convey("Registered", func() {
				driver := NewMockScheduler()
				driver.On("ReconcileTasks").Return("ok").Once()
				fID := mesos.FrameworkID{Value: proto.String("1234")}
				mInfo := mesos.MasterInfo{}
				s.Registered(driver, &fID, &mInfo)
				So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
			})

			Convey("Reregistered", func() {
				driver := NewMockScheduler()
				driver.On("ReconcileTasks").Return("ok").Once()
				db.Clean()
				s.Reregistered(driver, &mesos.MasterInfo{})
				So(driver.AssertCalled(t, "ReconcileTasks"), ShouldBeTrue)
			})

			Convey("Disconnected", func() {
				s.Disconnected(nil)
			})

			Convey("ResourceOffers", func() {
				driver := NewMockScheduler()

				Convey("No offers", func() {
					offers := []*mesos.Offer{}
					s.ResourceOffers(driver, offers)
					So(driver.AssertNotCalled(t, "DeclineOffer"), ShouldBeTrue)
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})

				Convey("No tasks", func() {
					offers := []*mesos.Offer{
						offer("1234", 1.0, 128),
					}
					driver.On("DeclineOffer").Return("declined").Once()
					s.ResourceOffers(driver, offers)
					So(driver.AssertCalled(t, "DeclineOffer"), ShouldBeTrue)
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})

				Convey("One task able to launch", func() {
					offers := []*mesos.Offer{
						offer("1234", 1.0, 128),
					}
					driver.On("LaunchTasks").Return("launched").Once()

					_, err := s.ScheduleTask(types.Request{
						TaskCPUs:    0.5,
						TaskMem:     22.0,
						DockerImage: "busybox",
						Command:     "echo hello",
					})

					So(err, ShouldBeNil)
					s.ResourceOffers(driver, offers)

					So(driver.AssertNotCalled(t, "DeclineOffer"), ShouldBeTrue)
					So(driver.AssertCalled(t, "LaunchTasks"), ShouldBeTrue)
				})

				Convey("One task unable to launch", func() {
					offers := []*mesos.Offer{
						offer("1234", 1.0, 128),
					}
					driver.On("DeclineOffer").Return("declined").Once()

					_, err := s.ScheduleTask(types.Request{
						TaskCPUs:    1.5,
						TaskMem:     22.0,
						DockerImage: "busybox",
						Command:     "echo hello",
					})

					So(err, ShouldBeNil)
					s.ResourceOffers(driver, offers)

					So(driver.AssertCalled(t, "DeclineOffer"), ShouldBeTrue)
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})
			})

			Convey("StatusUpdate", func() {
				Convey("Running then failing", func() {
					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_RUNNING.Enum(),
					})
					task, _ := db.ReadTask(id)
					So(len(task.Status), ShouldEqual, 1)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_RUNNING.String())

					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_FAILED.Enum(),
					})
					task, _ = db.ReadTask(id)

					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_RUNNING.String())
					So(task.Status[1].Status, ShouldEqual, mesos.TaskState_TASK_FAILED.String())
				})

				Convey("Failing immediately", func() {
					s.tasks = make(chan string, 100)
					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_FAILED.Enum(),
					})
					task, _ := db.ReadTask(id)
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_FAILED.String())
					So(task.Status[1].Status, ShouldEqual, mesos.TaskState_TASK_STAGING.String())

					select {
					case c := <-s.tasks:
						So(c, ShouldEqual, id)
					}
				})

				Convey("Task finished callback", func() {
					cb, ts := callbackReceiver()
					defer ts.Close()

					id := "eremetic-task.1000"
					db.PutTask(&types.EremeticTask{
						ID:          id,
						CallbackURI: ts.URL,
					})
					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_FINISHED.Enum(),
					})

					select {
					case c := <-cb:
						So(c.TaskID, ShouldEqual, id)
						So(c.Status, ShouldEqual, "TASK_FINISHED")
					}
				})

				Convey("Task failed callback", func() {
					cb, ts := callbackReceiver()
					defer ts.Close()

					id := "eremetic-task.1001"
					db.PutTask(&types.EremeticTask{
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

					select {
					case c := <-cb:
						So(c.TaskID, ShouldEqual, id)
						So(c.Status, ShouldEqual, "TASK_FAILED")
					}
				})

				Convey("Task retry callback", func() {
					cb, ts := callbackReceiver()
					defer ts.Close()

					id := "eremetic-task.1002"
					db.PutTask(&types.EremeticTask{
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

					select {
					case c := <-cb:
						So(c.TaskID, ShouldEqual, id)
						So(c.Status, ShouldEqual, "TASK_FINISHED")
					}
				})
			})

			Convey("FrameworkMessage", func() {
				driver := NewMockScheduler()
				message := `{"message": "this is a message"}`
				Convey("From Eremetic", func() {
					source := "eremetic-executor"
					executor := mesos.ExecutorID{
						Value: proto.String(source),
					}
					s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, message)
				})

				Convey("From an unknown source", func() {
					source := "other-source"
					executor := mesos.ExecutorID{
						Value: proto.String(source),
					}
					s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, message)
				})

				Convey("A bad json", func() {
					source := "eremetic-executor"
					executor := mesos.ExecutorID{
						Value: proto.String(source),
					}
					s.FrameworkMessage(driver, &executor, &mesos.SlaveID{}, "not a json")
				})
			})

			Convey("OfferRescinded", func() {
				s.OfferRescinded(nil, &mesos.OfferID{})
			})

			Convey("SlaveLost", func() {
				s.SlaveLost(nil, &mesos.SlaveID{})
			})

			Convey("ExecutorLost", func() {
				s.ExecutorLost(nil, &mesos.ExecutorID{}, &mesos.SlaveID{}, 2)
			})

			Convey("Error", func() {
				s.Error(nil, "Error")
			})
		})
	})

	Convey("ScheduleTask", t, func() {
		Convey("Given a valid Request", func() {
			scheduler := &eremeticScheduler{
				tasks:    make(chan string, 100),
				database: db,
			}

			request := types.Request{
				TaskCPUs:    0.5,
				TaskMem:     22.0,
				DockerImage: "busybox",
				Command:     "echo hello",
			}

			Convey("It should put a task id on the channel", func() {
				taskID, err := scheduler.ScheduleTask(request)

				So(err, ShouldBeNil)

				select {
				case c := <-scheduler.tasks:
					So(c, ShouldEqual, taskID)
					task, _ := db.ReadTask(taskID)
					So(task.TaskCPUs, ShouldEqual, request.TaskCPUs)
					So(task.TaskMem, ShouldEqual, request.TaskMem)
					So(task.Command, ShouldEqual, request.Command)
					So(task.User, ShouldEqual, "root")
					So(task.Environment, ShouldBeEmpty)
					So(task.Image, ShouldEqual, request.DockerImage)
					So(task.ID, ShouldStartWith, "eremetic-task.")
				}
			})
		})

		Convey("When the queue channel is full", func() {
			scheduler := &eremeticScheduler{
				tasks:    make(chan string, 1),
				database: db,
			}
			scheduler.tasks <- "dummy"

			request := types.Request{
				TaskCPUs:    0.5,
				TaskMem:     22.0,
				DockerImage: "busybox",
				Command:     "echo hello",
			}

			Convey("It should return an error", func() {
				_, err := scheduler.ScheduleTask(request)

				So(err, ShouldNotBeNil)
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
