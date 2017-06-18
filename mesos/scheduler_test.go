package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"github.com/eremetic-framework/eremetic/mock"
)

func callbackReceiver() (chan eremetic.CallbackData, *httptest.Server) {
	cb := make(chan eremetic.CallbackData, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var callback eremetic.CallbackData
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

	db := eremetic.NewDefaultTaskDB()

	Convey("Scheduling", t, func() {
		Convey("Given a scheduler with one task", func() {
			s := &Scheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			Convey("When creating a new scheduler", func() {
				queueSize := 200
				s = NewScheduler(&Settings{MaxQueueSize: queueSize}, db)

				Convey("The settings should have default values", func() {
					So(s.tasksCreated, ShouldEqual, 0)
					So(cap(s.tasks), ShouldEqual, queueSize)
				})
			})

			Convey("When the scheduler is registered", func() {
				driver := mock.NewMesosScheduler()
				driver.ReconcileTasksFn = func(ts []*mesosproto.TaskStatus) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}

				fid := mesosproto.FrameworkID{Value: proto.String("1234")}
				info := mesosproto.MasterInfo{}

				s.Registered(driver, &fid, &info)

				Convey("The tasks should be reconciled", func() {
					So(driver.ReconcileTasksFnInvoked, ShouldBeTrue)
				})

				Convey("The framework ID is stored", func() {
					So(s.frameworkID, ShouldEqual, "1234")
				})
			})

			Convey("When the scheduler is reregistered", func() {
				driver := mock.NewMesosScheduler()
				driver.ReconcileTasksFn = func(ts []*mesosproto.TaskStatus) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}
				db.Clean()

				s.Reregistered(driver, &mesosproto.MasterInfo{})

				Convey("The tasks should be reconciled", func() {
					So(driver.ReconcileTasksFnInvoked, ShouldBeTrue)
				})
			})

			Convey("When the scheduler is disconnected", func() {
				s.Disconnected(nil)
			})

			Convey("When an offer rescinded", func() {
				s.OfferRescinded(nil, &mesosproto.OfferID{})
			})

			Convey("When a slave was lost", func() {
				s.SlaveLost(nil, &mesosproto.SlaveID{})
			})

			Convey("When an executor was lost", func() {
				s.ExecutorLost(nil, &mesosproto.ExecutorID{}, &mesosproto.SlaveID{}, 2)
			})

			Convey("When there was an error", func() {
				s.Error(nil, "Error")
			})
		})
	})
	Convey("ResourceOffers", t, func() {
		Convey("Given a scheduler with one task", func() {
			s := &Scheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			driver := mock.NewMesosScheduler()

			Convey("When there are no offers", func() {
				offers := []*mesosproto.Offer{}
				s.ResourceOffers(driver, offers)

				Convey("No offers should be declined", func() {
					So(driver.DeclineOfferFnInvoked, ShouldBeFalse)
				})
				Convey("No tasks should be launched", func() {
					So(driver.LaunchTasksFnInvoked, ShouldBeFalse)
				})
			})
			Convey("When there are no tasks", func() {
				offers := []*mesosproto.Offer{
					offer("1234", 1.0, 128, &mesosproto.Unavailability{}),
				}
				driver.DeclineOfferFn = func(_ *mesosproto.OfferID, _ *mesosproto.Filters) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}

				s.ResourceOffers(driver, offers)

				Convey("The offer should be declined", func() {
					So(driver.DeclineOfferFnInvoked, ShouldBeTrue)
				})
				Convey("No tasks should be launched", func() {
					So(driver.LaunchTasksFnInvoked, ShouldBeFalse)
				})
			})
			Convey("When a task is able to launch", func() {
				offers := []*mesosproto.Offer{
					offer("1234", 1.0, 128, &mesosproto.Unavailability{}),
				}
				driver.LaunchTasksFn = func(_ []*mesosproto.OfferID, _ []*mesosproto.TaskInfo, _ *mesosproto.Filters) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}

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
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskQueued)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskStaging)
				})
				Convey("The offer should not be declined", func() {
					So(driver.DeclineOfferFnInvoked, ShouldBeFalse)
				})
				Convey("The tasks should be launched", func() {
					So(driver.LaunchTasksFnInvoked, ShouldBeTrue)
				})
			})

			Convey("When a task unable to launch", func() {
				offers := []*mesosproto.Offer{
					offer("1234", 1.0, 128, &mesosproto.Unavailability{}),
				}
				driver.DeclineOfferFn = func(_ *mesosproto.OfferID, _ *mesosproto.Filters) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}

				_, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    1.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(driver, offers)

				Convey("The offer should be declined", func() {
					So(driver.DeclineOfferFnInvoked, ShouldBeTrue)
				})
				Convey("The tasks should not be launched", func() {
					So(driver.LaunchTasksFnInvoked, ShouldBeFalse)
				})
			})

			Convey("When a task is marked for termination", func() {
				offers := []*mesosproto.Offer{offer("1234", 1.0, 128, &mesosproto.Unavailability{})}
				driver.DeclineOfferFn = func(_ *mesosproto.OfferID, _ *mesosproto.Filters) (mesosproto.Status, error) {
					return mesosproto.Status_DRIVER_RUNNING, nil
				}

				id, _ := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    1.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				task, _ := db.ReadTask(id)
				task.UpdateStatus(eremetic.Status{
					Time:   time.Now().Unix(),
					Status: eremetic.TaskTerminating,
				})
				db.PutTask(&task)

				s.ResourceOffers(driver, offers)

				task, _ = db.ReadTask(id)
				So(task.CurrentStatus(), ShouldEqual, eremetic.TaskKilled)
				So(driver.LaunchTasksFnInvoked, ShouldBeFalse)
				So(driver.DeclineOfferFnInvoked, ShouldBeTrue)
			})
		})
	})
	Convey("StatusUpdate", t, func() {
		Convey("Given a scheduler with one task", func() {
			s := &Scheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			Convey("When a running task fails", func() {
				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_RUNNING.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				So(len(task.Status), ShouldEqual, 1)
				So(task.Status[0].Status, ShouldEqual, eremetic.TaskRunning)

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FAILED.Enum(),
				})

				task, err = db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("The task status history should contain the failed status", func() {
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskRunning)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskFailed)
				})
			})

			Convey("When a task fails immediately", func() {
				s.tasks = make(chan string, 100)

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FAILED.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("The task should have a queued status", func() {
					So(len(task.Status), ShouldEqual, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskFailed)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskQueued)
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

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FINISHED.Enum(),
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

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FAILED.Enum(),
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

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FAILED.Enum(),
				})

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					State: mesosproto.TaskState_TASK_FINISHED.Enum(),
				})

				Convey("The callback data should be available", func() {
					c := <-cb

					So(c.TaskID, ShouldEqual, id)
					So(c.Status, ShouldEqual, "TASK_FINISHED")
				})
			})

			Convey("When the sandbox is updated", func() {
				id := "eremetic-task.1003"

				s.StatusUpdate(nil, &mesosproto.TaskStatus{
					TaskId: &mesosproto.TaskID{
						Value: proto.String(id),
					},
					Data:  []byte(`[{"Mounts":[{"Source":"/tmp/mesos/slaves/<agent_id>/frameworks/<framework_id>/executors/<task_id>/runs/<container_id>","Destination":"/mnt/mesos/sandbox","Mode":"","RW":true}]}]`),
					State: mesosproto.TaskState_TASK_RUNNING.Enum(),
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
		Convey("Given a scheduler with one task", func() {
			s := &Scheduler{
				tasks:    make(chan string, 1),
				database: db,
			}

			id := "eremetic-task.9999"
			db.PutTask(&eremetic.Task{ID: id})

			driver := mock.NewMesosScheduler()
			message := `{"message": "this is a message"}`

			Convey("When receiving a framework message from Eremetic", func() {
				source := "eremetic-executor"
				executor := mesosproto.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesosproto.SlaveID{}, message)
			})

			Convey("When receiving a framework message from an unknown source", func() {
				source := "other-source"
				executor := mesosproto.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesosproto.SlaveID{}, message)
			})

			Convey("When the framework message format is invalid", func() {
				source := "eremetic-executor"
				executor := mesosproto.ExecutorID{
					Value: proto.String(source),
				}
				s.FrameworkMessage(driver, &executor, &mesosproto.SlaveID{}, "not a json")
			})
		})
	})
	Convey("ScheduleTask", t, func() {
		Convey("Given a scheduler with no scheduled tasks", func() {
			scheduler := &Scheduler{
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
			scheduler := &Scheduler{
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

	Convey("KillTask", t, func() {
		driver := mock.NewMesosScheduler()
		id := "eremetic-task.9999"

		scheduler := &Scheduler{
			tasks:    make(chan string, 1),
			database: db,
			driver:   driver,
		}

		Convey("Given a running task", func() {
			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskRunning,
					},
				},
			})
			driver.KillTaskFn = func(_ *mesosproto.TaskID) (mesosproto.Status, error) {
				return mesosproto.Status_DRIVER_RUNNING, nil
			}

			err := scheduler.Kill(id)
			So(err, ShouldBeNil)
			So(driver.KillTaskFnInvoked, ShouldBeTrue)

			task, _ := db.ReadTask(id)
			So(task.CurrentStatus(), ShouldEqual, eremetic.TaskTerminating)
		})

		Convey("Given a queued task", func() {
			driver.KillTaskFn = func(_ *mesosproto.TaskID) (mesosproto.Status, error) {
				return mesosproto.Status_DRIVER_RUNNING, nil
			}
			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskQueued,
					},
				},
			})
			err := scheduler.Kill(id)
			So(err, ShouldBeNil)
			So(driver.KillTaskFnInvoked, ShouldBeFalse)
		})

		Convey("Given that something goes wrong", func() {
			driver.KillTaskFn = func(_ *mesosproto.TaskID) (mesosproto.Status, error) {
				return mesosproto.Status_DRIVER_RUNNING, errors.New("Nope")
			}

			err := scheduler.Kill(id)

			So(driver.KillTaskFnInvoked, ShouldBeTrue)
			So(err, ShouldNotBeNil)
		})

		Convey("Given already terminated task", func() {
			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskLost,
					},
				},
			})
			driver.KillTaskFn = func(_ *mesosproto.TaskID) (mesosproto.Status, error) {
				return mesosproto.Status_DRIVER_RUNNING, nil
			}

			err := scheduler.Kill(id)

			So(err, ShouldNotBeNil)
			So(driver.KillTaskFnInvoked, ShouldBeFalse)
		})
	})
}
