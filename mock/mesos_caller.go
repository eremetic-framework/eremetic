package mock

import (
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler"
)

type Caller struct {
	CallFn        func(call *scheduler.Call) (mesos.Response, error)
	CallFnInvoked bool
	Calls         []*scheduler.Call
}

func NewCaller() *Caller {
	return &Caller{}
}

func (m *Caller) Call(call *scheduler.Call) (mesos.Response, error) {
	m.CallFnInvoked = true
	m.Calls = append(m.Calls, call)
	return m.CallFn(call)
}
