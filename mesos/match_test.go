package mesos

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	"github.com/mesos/mesos-go/api/v0/mesosutil"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/eremetic-framework/eremetic"
	"time"
)

// Optional attributes can be added.
func offer(id string, cpu float64, mem float64, unavailability *mesosproto.Unavailability, attributes ...*mesosproto.Attribute) *mesosproto.Offer {
	return &mesosproto.Offer{
		Id: &mesosproto.OfferID{
			Value: proto.String(id),
		},
		FrameworkId: &mesosproto.FrameworkID{
			Value: proto.String("framework-1234"),
		},
		SlaveId: &mesosproto.SlaveID{
			Value: proto.String("slave-1234"),
		},
		Hostname: proto.String("localhost"),
		Resources: []*mesosproto.Resource{
			mesosutil.NewScalarResource("cpus", cpu),
			mesosutil.NewScalarResource("mem", mem),
		},
		Attributes:     attributes,
		Unavailability: unavailability,
	}
}

func TestMatch(t *testing.T) {
	offerA := offer("offer-a", 0.6, 200.0,
		&mesosproto.Unavailability{},
		&mesosproto.Attribute{
			Name: proto.String("role"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("badassmofo"),
			},
		},
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node1"),
			},
		},
	)
	offerB := offer("offer-b", 1.8, 512.0,
		&mesosproto.Unavailability{},
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node2"),
			},
		},
	)
	offerC := offer("offer-c", 1.8, 512.0,
		&mesosproto.Unavailability{
			Start: &mesosproto.TimeInfo{
				Nanoseconds: proto.Int64(time.Now().UnixNano()),
			},
			Duration: &mesosproto.DurationInfo{
				Nanoseconds: proto.Int64(time.Unix(0, 0).Add(1 * time.Hour).UnixNano()),
			},
		},
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node3"),
			},
		},
	)
	offerD := offer("offer-d", 1.8, 512.0,
		&mesosproto.Unavailability{
			Start: &mesosproto.TimeInfo{
				Nanoseconds: proto.Int64(time.Now().UnixNano()),
			},
		},
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node4"),
			},
		},
	)
	offerE := offer("offer-e", 1.8, 512.0,
		&mesosproto.Unavailability{
			Start: &mesosproto.TimeInfo{
				Nanoseconds: proto.Int64(time.Now().Add(-2 * time.Hour).UnixNano()),
			},
			Duration: &mesosproto.DurationInfo{
				Nanoseconds: proto.Int64(time.Unix(0, 0).Add(1 * time.Hour).UnixNano()),
			},
		},
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node3"),
			},
		},
	)

	Convey("CPUAvailable", t, func() {
		Convey("Above", func() {
			m := cpuAvailable(0.4)
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})

		Convey("Below", func() {
			m := cpuAvailable(0.8)
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("MemoryAvailable", t, func() {
		Convey("Above", func() {

			m := memoryAvailable(128.0)
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})

		Convey("Below", func() {
			m := memoryAvailable(256.0)
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Maintenance node", t, func() {
		Convey("Does not match (Defined maintenence window)", func() {
			m := availabilityMatch(time.Now())
			err := m.Matches(offerC)
			So(err, ShouldNotBeNil)
		})

		Convey("Does not match (Undefined maintenence window)", func() {
			m := availabilityMatch(time.Now())
			err := m.Matches(offerD)
			So(err, ShouldNotBeNil)
		})

		Convey("Does match (Maintenence window in past)", func() {
			m := availabilityMatch(time.Now())
			err := m.Matches(offerE)
			So(err, ShouldBeNil)
		})

		Convey("Does match the offer (task match)", func() {
			task := eremetic.Task{
				TaskCPUs: 0.6,
				TaskMem:  128.0,
			}
			offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerC, offerD})
			So(offer, ShouldEqual, offerA)
			So(others, ShouldHaveLength, 2)
		})
	})

	Convey("AttributeMatch", t, func() {
		Convey("Does match", func() {
			m := attributeMatch([]eremetic.AgentConstraint{
				eremetic.AgentConstraint{
					AttributeName:  "node_name",
					AttributeValue: "node1",
				},
			})
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})
		Convey("Does not match", func() {
			m := attributeMatch([]eremetic.AgentConstraint{
				eremetic.AgentConstraint{
					AttributeName:  "node_name",
					AttributeValue: "node2",
				},
			})
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("matchOffer", t, func() {
		Convey("Tasks without AgentConstraints", func() {
			Convey("Match", func() {
				task := eremetic.Task{
					TaskCPUs: 0.8,
					TaskMem:  128.0,
				}
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerB})

				So(offer, ShouldEqual, offerB)
				So(others, ShouldHaveLength, 1)
				So(others, ShouldContain, offerA)
			})

			Convey("No match CPU", func() {
				task := eremetic.Task{
					TaskCPUs: 2.0,
					TaskMem:  128.0,
				}
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})

			Convey("No match MEM", func() {
				task := eremetic.Task{
					TaskCPUs: 0.2,
					TaskMem:  712.0,
				}
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})
		})

		Convey("Tasks with AgentConstraints", func() {
			Convey("Match agent with attribute", func() {
				// Use task/mem constraints which match both offers.
				task := eremetic.Task{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					AgentConstraints: []eremetic.AgentConstraint{
						eremetic.AgentConstraint{
							AttributeName:  "node_name",
							AttributeValue: "node2",
						},
					},
				}
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerB})

				So(offer, ShouldEqual, offerB)
				So(others, ShouldHaveLength, 1)
				So(others, ShouldContain, offerA)
			})

			Convey("No matching slave with attribute", func() {
				// Use task/mem constraints which match both offers.
				task := eremetic.Task{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					AgentConstraints: []eremetic.AgentConstraint{
						eremetic.AgentConstraint{
							AttributeName:  "node_name",
							AttributeValue: "sherah",
						},
					},
				}
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})

			Convey("Match slave with mulitple attributes", func() {
				// Build two new offers, both with the same role as offerA.
				offerC := offer("offer-c", 0.6, 200.0,
					&mesosproto.Unavailability{},
					&mesosproto.Attribute{Name: proto.String("role"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("badassmofo")}},
					&mesosproto.Attribute{Name: proto.String("node_name"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("node3")}},
				)
				offerD := offer("offer-d", 0.6, 200.0,
					&mesosproto.Unavailability{},
					&mesosproto.Attribute{Name: proto.String("role"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("badassmofo")}},
				)

				task := eremetic.Task{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					AgentConstraints: []eremetic.AgentConstraint{
						eremetic.AgentConstraint{
							AttributeName:  "role",
							AttributeValue: "badassmofo",
						},
						eremetic.AgentConstraint{
							AttributeName:  "node_name",
							AttributeValue: "node3",
						},
					},
				}
				// Specifically add C last, our expected, so that we ensure
				// the other mocks do not match first.
				offer, others := matchOffer(task, []*mesosproto.Offer{offerA, offerD, offerC})

				So(offer, ShouldEqual, offerC)
				So(others, ShouldHaveLength, 2)
				So(others, ShouldContain, offerA)
				So(others, ShouldContain, offerD)
			})
		})
	})
}
