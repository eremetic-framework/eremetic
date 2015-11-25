package scheduler

import (
	"errors"
	"fmt"

	"github.com/alde/eremetic/types"
	log "github.com/dmuth/google-go-log4go"
	ogle "github.com/jacobsa/oglematchers"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type resourceMatcher struct {
	name  string
	value float64
}

func (m *resourceMatcher) Matches(o interface{}) error {
	offer := o.(*mesos.Offer)
	err := errors.New("")

	for _, res := range offer.Resources {
		if res.GetName() == m.name {
			if res.GetType() != mesos.Value_SCALAR {
				return err
			}

			if res.Scalar.GetValue() >= m.value {
				return nil
			}

			return err
		}
	}
	return err
}

func (m *resourceMatcher) Description() string {
	return fmt.Sprintf("%f of scalar resource %s", m.value, m.name)
}

func CPUAvailable(v float64) ogle.Matcher {
	return &resourceMatcher{"cpus", v}
}

func MemoryAvailable(v float64) ogle.Matcher {
	return &resourceMatcher{"mem", v}
}

func createMatcher(task types.EremeticTask) ogle.Matcher {
	return ogle.AllOf(
		CPUAvailable(task.TaskCPUs),
		MemoryAvailable(task.TaskMem))
}

func matches(matcher ogle.Matcher, o interface{}) bool {
	err := matcher.Matches(o)
	return err == nil
}

func matchOffer(task types.EremeticTask, offers []*mesos.Offer) (*mesos.Offer, []*mesos.Offer) {
	var matcher = createMatcher(task)
	for i, off := range offers {
		if matches(matcher, off) {
			offers[i] = offers[len(offers)-1]
			offers = offers[:len(offers)-1]
			return off, offers
		} else {
			log.Debugf("%s does not match: %s", off.Id.GetValue(), matcher.Description())
		}
	}
	return nil, offers
}
