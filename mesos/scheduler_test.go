package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler"

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

/*
func taskStatus(taskID string, state mesos.TaskState) mesos.TaskStatus {
	return mesos.TaskStatus{
		TaskID: &mesos.TaskID{
			Value: &taskID,
		},
		State: state,
	}
}
*/

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
					So(cap(s.tasks), ShouldEqual, queueSize)
				})
			})

			Convey("When the scheduler is subscribed", func() {
				caller := mock.NewCaller()
				caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
					return nil, nil
				}
				s.caller = caller

				fid := &mesos.FrameworkID{Value: "1234"}

				s.Subscribed(fid)

				Convey("The tasks should be reconciled", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_RECONCILE)
				})

				Convey("The framework ID is stored", func() {
					So(s.frameworkID, ShouldEqual, "1234")
				})
			})

			Convey("When the scheduler is re-subscribed", func() {
				// FIXME: what is this? call susbscribe twice?
				/*
					driver := mock.NewMesosScheduler()
					driver.ReconcileTasksFn = func(ts []*mesosproto.TaskStatus) (mesosproto.Status, error) {
						return mesosproto.Status_DRIVER_RUNNING, nil
					}
					db.Clean()

					s.Reregistered(driver, &mesosproto.MasterInfo{})

					Convey("The tasks should be reconciled", func() {
						So(driver.ReconcileTasksFnInvoked, ShouldBeTrue)
					})
				*/
			})
		})
	})
	Convey("ResourceOffers", t, func() {
		Convey("Given a scheduler with one task", func() {
			s := &Scheduler{
				tasks:    make(chan string, 1),
				database: db,
				random:   rand.New(rand.NewSource(time.Now().Unix())),
			}

			id := "eremetic-task.9999"

			db.PutTask(&eremetic.Task{ID: id})

			caller := mock.NewCaller()
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				return nil, nil
			}
			s.caller = caller

			Convey("When there are no offers", func() {
				offers := []mesos.Offer{}
				s.ResourceOffers(offers)

				Convey("No calls should be made", func() {
					So(caller.CallFnInvoked, ShouldBeFalse)
				})
			})
			Convey("When there are no tasks", func() {
				offers := []mesos.Offer{
					offer("1234", 1.0, 128, nil),
				}
				s.ResourceOffers(offers)

				Convey("The offer should be declined", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls, ShouldHaveLength, 1)
					So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_ACCEPT)
				})
			})

			Convey("When a task is able to launch", func() {
				caller := mock.NewCaller()
				caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
					return nil, nil
				}
				s.caller = caller
				offers := []mesos.Offer{
					offer("1234", 1.0, 128, nil),
				}

				taskID, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(offers)

				task, err := db.ReadTask(taskID)
				So(err, ShouldBeNil)

				Convey("The task should contain the status history", func() {
					So(task.Status, ShouldHaveLength, 2)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskQueued)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskStaging)
				})
				Convey("The tasks should be launched", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls, ShouldHaveLength, 1)
					So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_ACCEPT)
				})
			})

			Convey("When a task can be launched but fails", func() {
				caller := mock.NewCaller()
				caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
					return nil, errors.New("Nope")
				}
				s.caller = caller
				offers := []mesos.Offer{
					offer("1234", 1.0, 128, nil),
				}

				taskID, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(offers)

				task, err := db.ReadTask(taskID)
				So(err, ShouldBeNil)

				Convey("The task should contain the status history", func() {
					So(task.Status, ShouldHaveLength, 3)
					So(task.Status[0].Status, ShouldEqual, eremetic.TaskQueued)
					So(task.Status[1].Status, ShouldEqual, eremetic.TaskStaging)
					So(task.Status[2].Status, ShouldEqual, eremetic.TaskError)
				})
				Convey("The tasks should be launched", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls, ShouldHaveLength, 1)
					So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_ACCEPT)
				})
			})

			Convey("When a task unable to launch", func() {
				caller := mock.NewCaller()
				caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
					return nil, nil
				}
				s.caller = caller
				offers := []mesos.Offer{
					offer("1234", 1.0, 128, &mesos.Unavailability{}),
				}

				_, err := s.ScheduleTask(eremetic.Request{
					TaskCPUs:    1.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				})
				So(err, ShouldBeNil)

				s.ResourceOffers(offers)

				Convey("The offer should be declined", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls, ShouldHaveLength, 1)
					So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_ACCEPT)
					accept := caller.Calls[0].GetAccept()
					So(accept.OfferIDs, ShouldHaveLength, 1)
					So(accept.Operations[0].Launch.TaskInfos, ShouldHaveLength, 0)
				})
			})

			Convey("When a task is marked for termination", func() {
				caller := mock.NewCaller()
				caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
					return nil, nil
				}
				s.caller = caller
				offers := []mesos.Offer{offer("1234", 1.0, 128, &mesos.Unavailability{})}

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

				s.ResourceOffers(offers)

				task, _ = db.ReadTask(id)
				Convey("The task should be marked as killed", func() {
					So(task.CurrentStatus(), ShouldEqual, eremetic.TaskKilled)
				})
				Convey("The offer should be declined", func() {
					So(caller.CallFnInvoked, ShouldBeTrue)
					So(caller.Calls, ShouldHaveLength, 1)
					accept := caller.Calls[0].GetAccept()
					So(accept.OfferIDs, ShouldHaveLength, 1)
					So(accept.Operations[0].Launch.TaskInfos, ShouldHaveLength, 0)
				})
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
				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_RUNNING.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				So(len(task.Status), ShouldEqual, 1)
				So(task.Status[0].Status, ShouldEqual, eremetic.TaskRunning)

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FAILED.Enum(),
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

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FAILED.Enum(),
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

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FINISHED.Enum(),
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

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FAILED.Enum(),
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

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FAILED.Enum(),
				})

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_RUNNING.Enum(),
				})

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					State: mesos.TASK_FINISHED.Enum(),
				})

				Convey("The callback data should be available", func() {
					c := <-cb

					So(c.TaskID, ShouldEqual, id)
					So(c.Status, ShouldEqual, "TASK_FINISHED")
				})
			})

			Convey("When the sandbox is updated", func() {
				id := "eremetic-task.1003"

				s.StatusUpdate(mesos.TaskStatus{
					TaskID: mesos.TaskID{
						Value: id,
					},
					Data:  []byte(`[{"Mounts":[{"Source":"/tmp/mesos/slaves/<agent_id>/frameworks/<framework_id>/executors/<task_id>/runs/<container_id>","Destination":"/mnt/mesos/sandbox","Mode":"","RW":true}]}]`),
					State: mesos.TASK_RUNNING.Enum(),
				})

				task, err := db.ReadTask(id)
				So(err, ShouldBeNil)

				Convey("There should be a path to the sandbox", func() {
					So(task.SandboxPath, ShouldNotBeEmpty)
				})
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

	Convey("Schedule task with name from request", t, func() {
		Convey("Given a scheduler with no scheduled tasks", func() {
			scheduler := &Scheduler{
				tasks:    make(chan string, 100),
				database: db,
			}

			Convey("When scheduling a task with no name", func() {
				request := eremetic.Request{
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				}

				taskID, _ := scheduler.ScheduleTask(request)
				Convey("The task should have a name", func() {
					task, err := db.ReadTask(taskID)
					So(err, ShouldBeNil)
					So(task.Name, ShouldNotBeEmpty)
				})
			})

			Convey("When scheduling a task with a name from request", func() {
				request := eremetic.Request{
					Name:        "foobar",
					TaskCPUs:    0.5,
					TaskMem:     22.0,
					DockerImage: "busybox",
					Command:     "echo hello",
				}

				taskID, _ := scheduler.ScheduleTask(request)
				Convey("The task should have the same name as in request", func() {
					task, err := db.ReadTask(taskID)
					So(err, ShouldBeNil)
					So(task.Name, ShouldEqual, "foobar")
				})
			})
		})
	})

	Convey("KillTask", t, func() {
		id := "eremetic-task.9999"

		s := &Scheduler{
			tasks:    make(chan string, 1),
			database: db,
		}

		Convey("Given a running task", func() {
			caller := mock.NewCaller()
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				return nil, nil
			}
			s.caller = caller
			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskRunning,
					},
				},
			})

			err := s.Kill(id)
			So(err, ShouldBeNil)
			So(caller.CallFnInvoked, ShouldBeTrue)
			So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_KILL)

			task, _ := db.ReadTask(id)
			So(task.CurrentStatus(), ShouldEqual, eremetic.TaskTerminating)
		})

		Convey("Given a queued task", func() {
			caller := mock.NewCaller()
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				return nil, nil
			}
			s.caller = caller
			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskQueued,
					},
				},
			})
			err := s.Kill(id)
			So(err, ShouldBeNil)
			So(caller.CallFnInvoked, ShouldBeFalse)
		})

		Convey("Given that something goes wrong", func() {
			caller := mock.NewCaller()
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				return nil, errors.New("Nope")
			}
			s.caller = caller

			err := s.Kill(id)

			So(caller.CallFnInvoked, ShouldBeTrue)
			So(caller.Calls[0].GetType(), ShouldEqual, scheduler.Call_KILL)
			So(err, ShouldNotBeNil)
		})

		Convey("Given already terminated task", func() {
			caller := mock.NewCaller()
			caller.CallFn = func(call *scheduler.Call) (mesos.Response, error) {
				return nil, errors.New("Nope")
			}
			s.caller = caller

			db.PutTask(&eremetic.Task{
				ID: id,
				Status: []eremetic.Status{
					eremetic.Status{
						Time:   123456,
						Status: eremetic.TaskLost,
					},
				},
			})

			err := s.Kill(id)

			So(err, ShouldNotBeNil)
			So(caller.CallFnInvoked, ShouldBeFalse)
		})
	})
}
