package scheduler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	ogle "github.com/jacobsa/oglematchers"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type resourceMatcher struct {
	name  string
	value float64
}

type attributeMatcher struct {
	SlaveConstraints []types.SlaveConstraint
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

func (m *attributeMatcher) Matches(o interface{}) (err error) {
	offer := o.(*mesos.Offer)
	matched := int(0)

	for _, constraint := range m.SlaveConstraints {
		for _, attr := range offer.Attributes {
			if attr.GetName() == constraint.AttributeName {
				if attr.GetType() != mesos.Value_TEXT ||
					attr.Text.GetValue() != constraint.AttributeValue {
					err = errors.New("")

					// Match all constraints, not just one.
					return
				}
				matched += 1
			}
		}
	}

	if matched != len(m.SlaveConstraints) {
		err = errors.New("")
	}
	return
}

func (m *attributeMatcher) Description() string {
	descriptions := []string{}
	for _, constraint := range m.SlaveConstraints {
		descriptions = append(descriptions,
			fmt.Sprintf("slave attribute constraint %s=%s",
				constraint.AttributeName,
				constraint.AttributeValue,
			),
		)
	}
	return strings.Join(descriptions, ", ")
}

func AttributeMatch(slaveConstraints []types.SlaveConstraint) ogle.Matcher {
	return &attributeMatcher{slaveConstraints}
}

func createMatcher(task types.EremeticTask) ogle.Matcher {
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

func matchOffer(task types.EremeticTask, offers []*mesos.Offer) (*mesos.Offer, []*mesos.Offer) {
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
