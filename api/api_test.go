package api

import (
	"reflect"
	"testing"

	"github.com/eremetic-framework/eremetic"
)

var task = eremetic.Task{
	TaskCPUs:          12,
	TaskMem:           255,
	Command:           "task.Command",
	Args:              []string{"task.Args"},
	User:              "task.User",
	Environment:       map[string]string{"key": "value"},
	MaskedEnvironment: map[string]string{"key2": "masked_value"},
	Labels:            map[string]string{"label1": "label_value"},
	Image:             "task.Image",
	Volumes:           []eremetic.Volume{},
	Ports:             []eremetic.Port{},
	Status:            []eremetic.Status{eremetic.Status{Status: eremetic.TaskFinished, Time: 0}},
	ID:                "task.ID",
	Name:              "task.Name",
	Network:           "task.Network",
	DNS:               "task.DNS",
	FrameworkID:       "task.FrameworkID",
	AgentID:           "task.AgentID",
	AgentConstraints:  []eremetic.AgentConstraint{},
	Hostname:          "task.Hostname",
	Retry:             5,
	CallbackURI:       "task.CallbackURI",
	SandboxPath:       "task.SandboxPath",
	AgentIP:           "task.AgentIP",
	AgentPort:         1234,
	ForcePullImage:    true,
	Privileged:        false,
	FetchURIs:         []eremetic.URI{},
}

func TestAPI_V0_TaskV0FromTask_TaskFromV0(t *testing.T) {
	t0 := TaskV0FromTask(&task)
	ta := TaskFromV0(&t0)
	if !reflect.DeepEqual(ta, task) {
		t.Fatalf("Invalid conversion.\nExpected:\t%+v\nActual:\t%+v", ta, task)
	}
}

func TestAPI_V1_TaskV1FromTask_TaskFromV1(t *testing.T) {
	t1 := TaskV1FromTask(&task)
	ta := TaskFromV1(&t1)
	if !reflect.DeepEqual(ta, task) {
		t.Fatalf("Invalid conversion.\nExpected:\t%+v\nActual:\t%+v", ta, task)
	}
}
