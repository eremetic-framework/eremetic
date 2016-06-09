package scheduler

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	. "github.com/smartystreets/goconvey/convey"
)

// Optional attributes can be added.
func offer(id string, cpu float64, mem float64, attributes ...*mesos.Attribute) *mesos.Offer {
	return &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String(id),
		},
		FrameworkId: &mesos.FrameworkID{
			Value: proto.String("framework-1234"),
		},
		SlaveId: &mesos.SlaveID{
			Value: proto.String("slave-1234"),
		},
		Hostname: proto.String("localhost"),
		Url: &mesos.URL{
			Address: &mesos.Address{
				Port: proto.Int32(5050),
			},
		},
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", cpu),
			mesosutil.NewScalarResource("mem", mem),
		},
		Attributes: attributes,
	}
}

func TestMatch(t *testing.T) {
	offerA := offer("offer-a", 0.6, 200.0,
		&mesos.Attribute{
			Name: proto.String("role"),
			Type: mesos.Value_TEXT.Enum(),
			Text: &mesos.Value_Text{
				Value: proto.String("badassmofo"),
			},
		},
		&mesos.Attribute{
			Name: proto.String("node_name"),
			Type: mesos.Value_TEXT.Enum(),
			Text: &mesos.Value_Text{
				Value: proto.String("node1"),
			},
		},
	)
	offerB := offer("offer-b", 1.8, 512.0,
		&mesos.Attribute{
			Name: proto.String("node_name"),
			Type: mesos.Value_TEXT.Enum(),
			Text: &mesos.Value_Text{
				Value: proto.String("node2"),
			},
		},
	)

	Convey("CPUAvailable", t, func() {
		Convey("Above", func() {
			m := CPUAvailable(0.4)
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})

		Convey("Below", func() {
			m := CPUAvailable(0.8)
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("MemoryAvailable", t, func() {
		Convey("Above", func() {

			m := MemoryAvailable(128.0)
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})

		Convey("Below", func() {
			m := MemoryAvailable(256.0)
			err := m.Matches(offerA)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("AttributeMatch", t, func() {
		Convey("Does match", func() {
			m := AttributeMatch([]types.SlaveConstraint{
				types.SlaveConstraint{
					AttributeName:  "node_name",
					AttributeValue: "node1",
				},
			})
			err := m.Matches(offerA)
			So(err, ShouldBeNil)
		})
		Convey("Does not match", func() {
			m := AttributeMatch([]types.SlaveConstraint{
				types.SlaveConstraint{
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
				task := types.EremeticTask{
					TaskCPUs: 0.8,
					TaskMem:  128.0,
				}
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerB})

				So(offer, ShouldEqual, offerB)
				So(others, ShouldHaveLength, 1)
				So(others, ShouldContain, offerA)
			})

			Convey("No match CPU", func() {
				task := types.EremeticTask{
					TaskCPUs: 2.0,
					TaskMem:  128.0,
				}
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})

			Convey("No match MEM", func() {
				task := types.EremeticTask{
					TaskCPUs: 0.2,
					TaskMem:  712.0,
				}
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})
		})

		Convey("Tasks with SlaveConstraints", func() {
			Convey("Match slave with attribute", func() {
				// Use task/mem constraints which match both offers.
				task := types.EremeticTask{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					SlaveConstraints: []types.SlaveConstraint{
						types.SlaveConstraint{
							AttributeName:  "node_name",
							AttributeValue: "node2",
						},
					},
				}
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerB})

				So(offer, ShouldEqual, offerB)
				So(others, ShouldHaveLength, 1)
				So(others, ShouldContain, offerA)
			})

			Convey("No matching slave with attribute", func() {
				// Use task/mem constraints which match both offers.
				task := types.EremeticTask{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					SlaveConstraints: []types.SlaveConstraint{
						types.SlaveConstraint{
							AttributeName:  "node_name",
							AttributeValue: "sherah",
						},
					},
				}
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerB})

				So(offer, ShouldBeNil)
				So(others, ShouldHaveLength, 2)
			})

			Convey("Match slave with mulitple attributes", func() {
				// Build two new offers, both with the same role as offerA.
				offerC := offer("offer-c", 0.6, 200.0,
					&mesos.Attribute{Name: proto.String("role"), Type: mesos.Value_TEXT.Enum(), Text: &mesos.Value_Text{Value: proto.String("badassmofo")}},
					&mesos.Attribute{Name: proto.String("node_name"), Type: mesos.Value_TEXT.Enum(), Text: &mesos.Value_Text{Value: proto.String("node3")}},
				)
				offerD := offer("offer-d", 0.6, 200.0,
					&mesos.Attribute{Name: proto.String("role"), Type: mesos.Value_TEXT.Enum(), Text: &mesos.Value_Text{Value: proto.String("badassmofo")}},
				)

				task := types.EremeticTask{
					TaskCPUs: 0.5,
					TaskMem:  128.0,
					SlaveConstraints: []types.SlaveConstraint{
						types.SlaveConstraint{
							AttributeName:  "role",
							AttributeValue: "badassmofo",
						},
						types.SlaveConstraint{
							AttributeName:  "node_name",
							AttributeValue: "node3",
						},
					},
				}
				// Specifically add C last, our expected, so that we ensure
				// the other mocks do not match first.
				offer, others := matchOffer(task, []*mesos.Offer{offerA, offerD, offerC})

				So(offer, ShouldEqual, offerC)
				So(others, ShouldHaveLength, 2)
				So(others, ShouldContain, offerA)
				So(others, ShouldContain, offerD)
			})
		})
	})
}
