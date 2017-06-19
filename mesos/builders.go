package mesos

import (
	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	"github.com/mesos/mesos-go/api/v0/mesosutil"
)

// Optional attributes can be added.
func offer(id string, cpu float64, mem float64, unavailability *mesosproto.Unavailability, extra ...interface{}) *mesosproto.Offer {
	attributes := []*mesosproto.Attribute{}
	resources := []*mesosproto.Resource{
		mesosutil.NewScalarResource("cpus", cpu),
		mesosutil.NewScalarResource("mem", mem),
	}
	for _, r := range extra {
		switch r.(type) {
		case *mesosproto.Attribute:
			attributes = append(attributes, r.(*mesosproto.Attribute))
		case *mesosproto.Resource:
			resources = append(resources, r.(*mesosproto.Resource))
		}
	}
	return &mesosproto.Offer{
		Id: &mesosproto.OfferID{
			Value: proto.String(id),
		},
		FrameworkId: &mesosproto.FrameworkID{
			Value: proto.String("framework-1234"),
		},
		SlaveId: &mesosproto.SlaveID{
			Value: proto.String("agent-id"),
		},
		Hostname:       proto.String("localhost"),
		Resources:      resources,
		Attributes:     attributes,
		Unavailability: unavailability,
	}
}

func textAttribute(name string, value string) *mesosproto.Attribute {
	return &mesosproto.Attribute{
		Name: proto.String(name),
		Type: mesosproto.Value_TEXT.Enum(),
		Text: &mesosproto.Value_Text{
			Value: proto.String(value),
		},
	}
}

func unavailability(details ...int64) *mesosproto.Unavailability {
	un := mesosproto.Unavailability{}
	if len(details) >= 1 {
		un.Start = &mesosproto.TimeInfo{
			Nanoseconds: proto.Int64(details[0]),
		}
	}
	if len(details) >= 2 {
		un.Duration = &mesosproto.DurationInfo{
			Nanoseconds: proto.Int64(details[1]),
		}
	}
	return &un
}
