package scheduler

import (
	"testing"

	"github.com/klarna/eremetic/types"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	. "github.com/smartystreets/goconvey/convey"
)

func offer(id string, cpu float64, mem float64) *mesos.Offer {
	return &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String(id),
		},
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", cpu),
			mesosutil.NewScalarResource("mem", mem),
		},
	}
}

func TestMatch(t *testing.T) {
	offerA := offer("offer-a", 0.6, 200.0)
	offerB := offer("offer-b", 1.8, 512.0)

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

	Convey("matchOffer", t, func() {
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
}
