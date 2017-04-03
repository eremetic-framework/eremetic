package mesos

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/mesosproto"
	"github.com/mesos/mesos-go/api/v0/mesosutil"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/klarna/eremetic"
)

// Optional attributes can be added.
func offer(id string, cpu float64, mem float64, attributes ...*mesosproto.Attribute) *mesosproto.Offer {
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
		Attributes: attributes,
	}
}

func TestMatch(t *testing.T) {
	offerA := offer("offer-a", 0.6, 200.0,
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
		&mesosproto.Attribute{
			Name: proto.String("node_name"),
			Type: mesosproto.Value_TEXT.Enum(),
			Text: &mesosproto.Value_Text{
				Value: proto.String("node2"),
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

	Convey("AttributeMatch", t, func() {
		Convey("Does match", func() {
			m := attributeMatch([]eremetic.SlaveConstraint{
				eremetic.SlaveConstraint{
					AttributeName:  "node_name",
					AttributeValue: "node1",
				},
			})
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})
		Convey("Does not match", func() {
			m := attributeMatch([]eremetic.SlaveConstraint{
				eremetic.SlaveConstraint{
					AttributeName:  "node_name",
					AttributeValue: "node2",
				},
			})
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("matchOffer", t, func() {
		Convey("Tasks without SlaveConstraints", func() {
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

		Convey("Tasks with SlaveConstraints", func() {
			Convey("Match slave with attribute", func() {
				// Use task/mem constraints which match both offers.
				task := eremetic.Task{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					SlaveConstraints: []eremetic.SlaveConstraint{
						eremetic.SlaveConstraint{
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
					SlaveConstraints: []eremetic.SlaveConstraint{
						eremetic.SlaveConstraint{
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
					&mesosproto.Attribute{Name: proto.String("role"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("badassmofo")}},
					&mesosproto.Attribute{Name: proto.String("node_name"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("node3")}},
				)
				offerD := offer("offer-d", 0.6, 200.0,
					&mesosproto.Attribute{Name: proto.String("role"), Type: mesosproto.Value_TEXT.Enum(), Text: &mesosproto.Value_Text{Value: proto.String("badassmofo")}},
				)

				task := eremetic.Task{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					SlaveConstraints: []eremetic.SlaveConstraint{
						eremetic.SlaveConstraint{
							AttributeName:  "role",
							AttributeValue: "badassmofo",
						},
						eremetic.SlaveConstraint{
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
