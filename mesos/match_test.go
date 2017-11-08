package mesos

import (
	"testing"

	"github.com/mesos/mesos-go/api/v0/mesosproto"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/cybricio/eremetic"
	"time"
)

func TestMatch(t *testing.T) {
	offerA := offer("offer-a", 0.6, 200.0,
		unavailability(),
		textAttribute("role", "badassmofo"),
		textAttribute("node_name", "node1"),
	)
	offerB := offer("offer-b", 1.8, 512.0,
		unavailability(),
		textAttribute("node_name", "node2"),
	)
	offerC := offer("offer-c", 1.8, 512.0,
		unavailability(
			time.Now().UnixNano(),
			time.Unix(0, 0).Add(1*time.Hour).UnixNano(),
		),
		textAttribute("node_name", "node3"),
	)
	offerD := offer("offer-d", 1.8, 512.0,
		unavailability(
			time.Now().UnixNano(),
		),
		textAttribute("node_name", "node4"),
	)
	offerE := offer("offer-e", 1.8, 512.0,
		unavailability(
			time.Now().Add(-2*time.Hour).UnixNano(),
			time.Unix(0, 0).Add(1*time.Hour).UnixNano(),
		),
		textAttribute("node_name", "node3"),
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
					unavailability(),
					textAttribute("role", "badassmofo"),
					textAttribute("node_name", "node3"),
				)
				offerD := offer("offer-d", 0.6, 200.0,
					unavailability(),
					textAttribute("role", "badassmofo"),
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
