package scheduler

import (
	"fmt"
	"os"
	"testing"

	"github.com/alde/eremetic/database"
	"github.com/alde/eremetic/types"
	log "github.com/dmuth/google-go-log4go"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestScheduler(t *testing.T) {
	dir, _ := os.Getwd()
	database.NewDB(fmt.Sprintf("%s/../db/test.db", dir))
	database.Clean()
	defer database.Close()

	Convey("eremeticScheduler", t, func() {
		s := eremeticScheduler{}
		id := "eremetic-task.9999"
		database.PutTask(&types.EremeticTask{ID: id})

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
			newTask := s.newTask(&offer, &task)

			So(newTask.GetTaskId().GetValue(), ShouldEqual, task.ID)
		})

		Convey("createEremeticScheduler", func() {
			s := createEremeticScheduler()
			So(s.tasksCreated, ShouldEqual, 0)
		})

		Convey("API", func() {
			Convey("Registered", func() {
				fID := mesos.FrameworkID{Value: proto.String("1234")}
				mInfo := mesos.MasterInfo{}
				s.Registered(nil, &fID, &mInfo)
			})

			Convey("Reregistered", func() {
				s.Reregistered(nil, &mesos.MasterInfo{})
			})

			Convey("Disconnected", func() {
				s.Disconnected(nil)
			})

			Convey("ResourceOffers", func() {
				driver := NewMockScheduler()
				var offers []*mesos.Offer

				Convey("No offers", func() {
					s.ResourceOffers(driver, offers)
					So(driver.AssertNotCalled(t, "DeclineOffer"), ShouldBeTrue)
					So(driver.AssertNotCalled(t, "LaunchTasks"), ShouldBeTrue)
				})

				Convey("No tasks", func() {
					offers = append(offers, &mesos.Offer{Id: &mesos.OfferID{Value: proto.String("1234")}})
					driver.On("DeclineOffer").Return("declined").Once()
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
					task, _ := database.ReadTask(id)
					So(len(task.Status), ShouldEqual, 1)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_RUNNING.String())

					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_FAILED.Enum(),
					})
					task, _ = database.ReadTask(id)

					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_RUNNING.String())
					So(task.Status[1].Status, ShouldEqual, mesos.TaskState_TASK_FAILED.String())
				})

				Convey("Failing immediatly", func() {
					s.tasks = make(chan string, 100)
					s.StatusUpdate(nil, &mesos.TaskStatus{
						TaskId: &mesos.TaskID{
							Value: proto.String(id),
						},
						State: mesos.TaskState_TASK_FAILED.Enum(),
					})
					task, _ := database.ReadTask(id)
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, mesos.TaskState_TASK_FAILED.String())
					So(task.Status[1].Status, ShouldEqual, mesos.TaskState_TASK_STAGING.String())

					select {
					case c := <-s.tasks:
						So(c, ShouldEqual, id)
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
				tasks: make(chan string, 100),
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
					task, _ := database.ReadTask(taskID)
					So(task.TaskCPUs, ShouldEqual, request.TaskCPUs)
					So(task.TaskMem, ShouldEqual, request.TaskMem)
					So(task.Command.GetValue(), ShouldEqual, request.Command)
					So(*task.Container.Docker.Image, ShouldEqual, request.DockerImage)
					So(task.ID, ShouldStartWith, "eremetic-task.")
				}
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

func (sched *MockScheduler) Error(d sched.SchedulerDriver, msg string) {
	log.Error(msg)
	sched.Called()
}
