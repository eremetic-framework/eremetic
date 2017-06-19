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
