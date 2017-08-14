package mesos

import (
	"github.com/mesos/mesos-go/api/v1/lib"
)

// Optional attributes can be added.
func offer(id string, cpu float64, mem float64, unavailability *mesos.Unavailability, extra ...interface{}) mesos.Offer {
	attributes := []mesos.Attribute{}
	resources := []mesos.Resource{
		*mesos.BuildResource().Name("cpus").Scalar(cpu).Resource,
		*mesos.BuildResource().Name("mem").Scalar(mem).Resource,
	}
	for _, r := range extra {
		switch r.(type) {
		case mesos.Attribute:
			attributes = append(attributes, r.(mesos.Attribute))
		case mesos.Resource:
			resources = append(resources, r.(mesos.Resource))
		}
	}
	return mesos.Offer{
		ID: mesos.OfferID{
			Value: id,
		},
		FrameworkID: mesos.FrameworkID{
			Value: "framework-1234",
		},
		AgentID: mesos.AgentID{
			Value: "agent-id",
		},
		Hostname:       "localhost",
		Resources:      resources,
		Attributes:     attributes,
		Unavailability: unavailability,
	}
}

func textAttribute(name string, value string) mesos.Attribute {
	return mesos.Attribute{
		Name: name,
		Type: mesos.TEXT,
		Text: &mesos.Value_Text{
			Value: value,
		},
	}
}

func unavailability(details ...int64) *mesos.Unavailability {
	un := mesos.Unavailability{}
	if len(details) >= 1 {
		un.Start = mesos.TimeInfo{
			Nanoseconds: details[0],
		}
	}
	if len(details) >= 2 {
		un.Duration = &mesos.DurationInfo{
			Nanoseconds: details[1],
		}
	}
	return &un
}
