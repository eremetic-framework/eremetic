package mesos

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	ogle "github.com/jacobsa/oglematchers"
	"github.com/mesos/mesos-go/api/v0/mesosproto"

	"github.com/klarna/eremetic"
)

type resourceMatcher struct {
	name  string
	value float64
}

type attributeMatcher struct {
	constraint eremetic.SlaveConstraint
}

func (m *resourceMatcher) Matches(o interface{}) error {
	offer := o.(*mesosproto.Offer)
	err := errors.New("")

	for _, res := range offer.Resources {
		if res.GetName() == m.name {
			if res.GetType() != mesosproto.Value_SCALAR {
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

func cpuAvailable(v float64) ogle.Matcher {
	return &resourceMatcher{"cpus", v}
}

func memoryAvailable(v float64) ogle.Matcher {
	return &resourceMatcher{"mem", v}
}

func (m *attributeMatcher) Matches(o interface{}) error {
	offer := o.(*mesosproto.Offer)

	for _, attr := range offer.Attributes {
		if attr.GetName() == m.constraint.AttributeName {
			if attr.GetType() != mesosproto.Value_TEXT ||
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

func attributeMatch(slaveConstraints []eremetic.SlaveConstraint) ogle.Matcher {
	var submatchers []ogle.Matcher
	for _, constraint := range slaveConstraints {
		submatchers = append(submatchers, &attributeMatcher{constraint})
	}
	return ogle.AllOf(submatchers...)
}

func createMatcher(task eremetic.Task) ogle.Matcher {
	return ogle.AllOf(
		cpuAvailable(task.TaskCPUs),
		memoryAvailable(task.TaskMem),
		attributeMatch(task.SlaveConstraints),
	)
}

func matches(matcher ogle.Matcher, o interface{}) bool {
	err := matcher.Matches(o)
	return err == nil
}

func matchOffer(task eremetic.Task, offers []*mesosproto.Offer) (*mesosproto.Offer, []*mesosproto.Offer) {
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
