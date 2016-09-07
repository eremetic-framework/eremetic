package scheduler

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	ogle "github.com/jacobsa/oglematchers"
	"github.com/klarna/eremetic"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type resourceMatcher struct {
	name  string
	value float64
}

type attributeMatcher struct {
	constraint eremetic.SlaveConstraint
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

func (m *attributeMatcher) Matches(o interface{}) error {
	offer := o.(*mesos.Offer)

	for _, attr := range offer.Attributes {
		if attr.GetName() == m.constraint.AttributeName {
			if attr.GetType() != mesos.Value_TEXT ||
				attr.Text.GetValue() != m.constraint.AttributeValue {
				return errors.New("")
			}
			return nil
		}
	}

	return errors.New("")
}

func (m *attributeMatcher) Description() string {
	return fmt.Sprintf("slave attribute constraint %s=%s",
		m.constraint.AttributeName,
		m.constraint.AttributeValue,
	)
}

func AttributeMatch(slaveConstraints []eremetic.SlaveConstraint) ogle.Matcher {
	var submatchers []ogle.Matcher
	for _, constraint := range slaveConstraints {
		submatchers = append(submatchers, &attributeMatcher{constraint})
	}
	return ogle.AllOf(submatchers...)
}

func createMatcher(task eremetic.Task) ogle.Matcher {
	return ogle.AllOf(
		CPUAvailable(task.TaskCPUs),
		MemoryAvailable(task.TaskMem),
		AttributeMatch(task.SlaveConstraints),
	)
}

func matches(matcher ogle.Matcher, o interface{}) bool {
	err := matcher.Matches(o)
	return err == nil
}

func matchOffer(task eremetic.Task, offers []*mesos.Offer) (*mesos.Offer, []*mesos.Offer) {
	var matcher = createMatcher(task)
	for i, off := range offers {
		if matches(matcher, off) {
			offers[i] = offers[len(offers)-1]
			offers = offers[:len(offers)-1]
			return off, offers
		}
		logrus.WithFields(logrus.Fields{
			"offer_id": off.Id.GetValue(),
			"matcher":  matcher.Description(),
			"task_id":  task.ID,
		}).Debug("Unable to match offer")
	}
	return nil, offers
}
