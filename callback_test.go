package eremetic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type callbackHandler struct {
	Invoked bool
	Payload map[string]interface{}
}

func (h *callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&h.Payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.Invoked = true
}

func TestCallback(t *testing.T) {
	Convey("Given an empty task", t, func() {
		var h callbackHandler
		ts := httptest.NewServer(&h)

		var task Task

		Convey("When notifying without a callback URI", func() {
			NotifyCallback(&task)
			time.Sleep(50 * time.Millisecond)

			Convey("The callback handler should not be invoked", func() {
				So(h.Invoked, ShouldBeFalse)
			})
		})
		Convey("When notifying without any statuses", func() {
			task.CallbackURI = ts.URL

			NotifyCallback(&task)
			time.Sleep(10 * time.Millisecond)

			Convey("The callback handler should be invoked", func() {
				So(h.Invoked, ShouldBeFalse)
			})
		})
		Convey("When notifying with one status", func() {
			task.CallbackURI = ts.URL
			task.Status = []Status{
				{Time: 0, Status: TaskState_TASK_STAGING},
			}

			NotifyCallback(&task)
			time.Sleep(10 * time.Millisecond)

			Convey("The callback handler should be invoked", func() {
				So(h.Invoked, ShouldBeTrue)
			})
			Convey("The callback payload status should contain the status", func() {
				So(h.Payload, ShouldContainKey, "status")
				So(h.Payload["status"], ShouldEqual, "TASK_STAGING")
			})
		})
		Convey("When notifying with many statuses", func() {
			task.CallbackURI = ts.URL
			task.Status = []Status{
				{Time: 0, Status: TaskState_TASK_STAGING},
				{Time: 1, Status: TaskState_TASK_RUNNING},
				{Time: 2, Status: TaskState_TASK_FINISHED},
			}

			NotifyCallback(&task)
			time.Sleep(10 * time.Millisecond)

			Convey("The callback handler should be invoked", func() {
				So(h.Invoked, ShouldBeTrue)
			})
			Convey("The callback payload status should contain the latest status", func() {
				So(h.Payload, ShouldContainKey, "status")
				So(h.Payload["status"], ShouldEqual, "TASK_FINISHED")
			})
		})
	})
}
