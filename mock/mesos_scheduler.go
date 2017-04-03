package mock

import (
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	"github.com/mesos/mesos-go/api/v0/scheduler"
)

// MesosScheduler implements a mocked mesos scheduler iterface for testing
type MesosScheduler struct {
	AbortFn                       func() (mesosproto.Status, error)
	AbortFnInvoked                bool
	AcceptOffersFn                func([]*mesosproto.OfferID, []*mesosproto.Offer_Operation, *mesosproto.Filters) (mesosproto.Status, error)
	AcceptOffersFnInvoked         bool
	DeclineOfferFn                func(*mesosproto.OfferID, *mesosproto.Filters) (mesosproto.Status, error)
	DeclineOfferFnInvoked         bool
	JoinFn                        func() (mesosproto.Status, error)
	JoinFnInvoked                 bool
	KillTaskFn                    func(*mesosproto.TaskID) (mesosproto.Status, error)
	KillTaskFnInvoked             bool
	ReconcileTasksFn              func([]*mesosproto.TaskStatus) (mesosproto.Status, error)
	ReconcileTasksFnInvoked       bool
	RequestResourcesFn            func([]*mesosproto.Request) (mesosproto.Status, error)
	RequestResourcesFnInvoked     bool
	ReviveOffersFn                func() (mesosproto.Status, error)
	ReviveOffersFnInvoked         bool
	RunFn                         func() (mesosproto.Status, error)
	RunFnInvoked                  bool
	StartFn                       func() (mesosproto.Status, error)
	StartFnInvoked                bool
	StopFn                        func(bool) (mesosproto.Status, error)
	StopFnInvoked                 bool
	SendFrameworkMessageFn        func(*mesosproto.ExecutorID, *mesosproto.SlaveID, string) (mesosproto.Status, error)
	SendFrameworkMessageFnInvoked bool
	LaunchTasksFn                 func([]*mesosproto.OfferID, []*mesosproto.TaskInfo, *mesosproto.Filters) (mesosproto.Status, error)
	LaunchTasksFnInvoked          bool
	RegisteredFn                  func(scheduler.SchedulerDriver, *mesosproto.FrameworkID, *mesosproto.MasterInfo)
	RegisteredFnInvoked           bool
	ReregisteredFn                func(scheduler.SchedulerDriver, *mesosproto.MasterInfo)
	ReregisteredFnInvoked         bool
	DisconnectedFn                func(scheduler.SchedulerDriver)
	DisconnectedFnInvoked         bool
	ResourceOffersFn              func(scheduler.SchedulerDriver, []*mesosproto.Offer)
	ResourceOffersFnInvoked       bool
	OfferRescindedFn              func(scheduler.SchedulerDriver, *mesosproto.OfferID)
	OfferRescindedFnInvoked       bool
	StatusUpdateFn                func(scheduler.SchedulerDriver, *mesosproto.TaskStatus)
	StatusUpdateFnInvoked         bool
	FrameworkMessageFn            func(scheduler.SchedulerDriver, *mesosproto.ExecutorID, *mesosproto.SlaveID, string)
	FrameworkMessageFnInvoked     bool
	SlaveLostFn                   func(scheduler.SchedulerDriver, *mesosproto.SlaveID)
	SlaveLostFnInvoked            bool
	ExecutorLostFn                func(scheduler.SchedulerDriver, *mesosproto.ExecutorID, *mesosproto.SlaveID, int)
	ExecutorLostFnInvoked         bool
	ErrorFn                       func(scheduler.SchedulerDriver, string)
	ErrorFnInvoked                bool
}

// NewMesosScheduler returns a new mocked mesos scheduler
func NewMesosScheduler() *MesosScheduler {
	return &MesosScheduler{}
}

func (m *MesosScheduler) Abort() (stat mesosproto.Status, err error) {
	m.AbortFnInvoked = true
	return m.AbortFn()
}

func (m *MesosScheduler) AcceptOffers(offerIds []*mesosproto.OfferID, operations []*mesosproto.Offer_Operation, filters *mesosproto.Filters) (mesosproto.Status, error) {
	m.AcceptOffersFnInvoked = true
	return m.AcceptOffersFn(offerIds, operations, filters)
}

func (m *MesosScheduler) DeclineOffer(offerID *mesosproto.OfferID, filters *mesosproto.Filters) (mesosproto.Status, error) {
	m.DeclineOfferFnInvoked = true
	return m.DeclineOfferFn(offerID, filters)
}

func (m *MesosScheduler) Join() (mesosproto.Status, error) {
	m.JoinFnInvoked = true
	return m.JoinFn()
}

func (m *MesosScheduler) KillTask(id *mesosproto.TaskID) (mesosproto.Status, error) {
	m.KillTaskFnInvoked = true
	return m.KillTaskFn(id)
}

func (m *MesosScheduler) ReconcileTasks(ts []*mesosproto.TaskStatus) (mesosproto.Status, error) {
	m.ReconcileTasksFnInvoked = true
	return m.ReconcileTasksFn(ts)
}

func (m *MesosScheduler) RequestResources(r []*mesosproto.Request) (mesosproto.Status, error) {
	m.RequestResourcesFnInvoked = true
	return m.RequestResourcesFn(r)
}

func (m *MesosScheduler) ReviveOffers() (mesosproto.Status, error) {
	m.ReviveOffersFnInvoked = true
	return m.ReviveOffersFn()
}

func (m *MesosScheduler) Run() (mesosproto.Status, error) {
	m.RunFnInvoked = true
	return m.RunFn()
}

func (m *MesosScheduler) Start() (mesosproto.Status, error) {
	m.StartFnInvoked = true
	return m.StartFn()
}

func (m *MesosScheduler) Stop(b bool) (mesosproto.Status, error) {
	m.StopFnInvoked = true
	return m.StopFn(b)
}

func (m *MesosScheduler) SendFrameworkMessage(eID *mesosproto.ExecutorID, sID *mesosproto.SlaveID, s string) (mesosproto.Status, error) {
	m.SendFrameworkMessageFnInvoked = true
	return m.SendFrameworkMessageFn(eID, sID, s)
}

func (m *MesosScheduler) LaunchTasks(o []*mesosproto.OfferID, t []*mesosproto.TaskInfo, f *mesosproto.Filters) (mesosproto.Status, error) {
	m.LaunchTasksFnInvoked = true
	return m.LaunchTasksFn(o, t, f)
}

func (m *MesosScheduler) Registered(s scheduler.SchedulerDriver, f *mesosproto.FrameworkID, minfo *mesosproto.MasterInfo) {
	m.RegisteredFnInvoked = true
	m.RegisteredFn(s, f, minfo)
}

func (m *MesosScheduler) Reregistered(s scheduler.SchedulerDriver, info *mesosproto.MasterInfo) {
	m.ReregisteredFnInvoked = true
	m.ReregisteredFn(s, info)
}

func (m *MesosScheduler) Disconnected(s scheduler.SchedulerDriver) {
	m.DisconnectedFnInvoked = true
	m.DisconnectedFn(s)
}

func (m *MesosScheduler) ResourceOffers(s scheduler.SchedulerDriver, o []*mesosproto.Offer) {
	m.ResourceOffersFnInvoked = true
	m.ResourceOffersFn(s, o)
}

func (m *MesosScheduler) OfferRescinded(s scheduler.SchedulerDriver, o *mesosproto.OfferID) {
	m.OfferRescindedFnInvoked = true
	m.OfferRescindedFn(s, o)
}

func (m *MesosScheduler) StatusUpdate(s scheduler.SchedulerDriver, ts *mesosproto.TaskStatus) {
	m.StatusUpdateFnInvoked = true
	m.StatusUpdateFn(s, ts)
}

func (m *MesosScheduler) FrameworkMessage(sd scheduler.SchedulerDriver, eID *mesosproto.ExecutorID, sID *mesosproto.SlaveID, s string) {
	m.FrameworkMessageFnInvoked = true
	m.FrameworkMessageFn(sd, eID, sID, s)
}

func (m *MesosScheduler) SlaveLost(s scheduler.SchedulerDriver, sID *mesosproto.SlaveID) {
	m.SlaveLostFnInvoked = true
	m.SlaveLostFn(s, sID)
}

func (m *MesosScheduler) ExecutorLost(sd scheduler.SchedulerDriver, eID *mesosproto.ExecutorID, sID *mesosproto.SlaveID, i int) {
	m.ExecutorLostFnInvoked = true
	m.ExecutorLostFn(sd, eID, sID, i)
}

func (m *MesosScheduler) Error(d scheduler.SchedulerDriver, msg string) {
	m.ErrorFnInvoked = true
	m.ErrorFn(d, msg)
}
