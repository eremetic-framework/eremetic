package mesos

import (
	"errors"
	"fmt"
	"strings"
	"strconv"

	"github.com/sirupsen/logrus"
	ogle "github.com/jacobsa/oglematchers"
	"github.com/mesos/mesos-go/api/v0/mesosproto"

        "sort"
	"time"

	"github.com/rockerbox/eremetic"
)

type resourceMatcher struct {
	name  string
	value float64
}

type attributeMatcher struct {
	constraint eremetic.AgentConstraint
}

type availabilityMatcher struct {
	time.Time
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

func availabilityMatch(matchTime time.Time) ogle.Matcher {
	return &availabilityMatcher{matchTime}
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

func (m *availabilityMatcher) Matches(o interface{}) error {
	offer := o.(*mesosproto.Offer)

	if offer.Unavailability == nil {
		return nil
	}

	if start := offer.Unavailability.GetStart(); start != nil && m.UnixNano() >= *start.Nanoseconds {
		if duration := offer.Unavailability.GetDuration(); duration == nil {
			return errors.New("node is on indefinite period of maintenance")
		} else if m.UnixNano() <= *start.Nanoseconds+*duration.Nanoseconds {
			return errors.New("node is currently in maintenance mode")
		}
	}
	return nil
}

func (m *availabilityMatcher) Description() string {
	return fmt.Sprintf("availability matcher")
}

func (m *attributeMatcher) Description() string {
	return fmt.Sprintf("agent attribute constraint %s=%s",
		m.constraint.AttributeName,
		m.constraint.AttributeValue,
	)
}

func attributeMatch(agentConstraints []eremetic.AgentConstraint) ogle.Matcher {
	var submatchers []ogle.Matcher
	for _, constraint := range agentConstraints {
		submatchers = append(submatchers, &attributeMatcher{constraint})
	}
	return ogle.AllOf(submatchers...)
}

func createMatcher(task eremetic.Task) ogle.Matcher {
	return ogle.AllOf(
		cpuAvailable(task.TaskCPUs),
		memoryAvailable(task.TaskMem),
		attributeMatch(task.AgentConstraints),
		availabilityMatch(time.Now()),
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
                        offers = append(offers[:i], offers[i+1:]...) // preserve sort order
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

// By is a helper to sort offer to best-fit bin pack by mem
type By func(o1, o2 *mesosproto.Offer) bool

// Sort provides sort function across offer array
func (by By) Sort(offers []*mesosproto.Offer) {
	os := &offerSorter{
		offers: offers,
		by:     by,
	}
	sort.Sort(os)
}

type offerSorter struct {
	offers []*mesosproto.Offer
	by     func(o1, o2 *mesosproto.Offer) bool
}

func (s *offerSorter) Len() int {
	return len(s.offers)
}

func (s *offerSorter) Swap(i, j int) {
	s.offers[i], s.offers[j] = s.offers[j], s.offers[i]
}

func (s *offerSorter) Less(i, j int) bool {
	return s.by(s.offers[i], s.offers[j])
}


func sortByLeastMemAvailable(offers []*mesosproto.Offer) {
	byID := func(o1, o2 *mesosproto.Offer) bool {
		s1 := o1.GetSlaveId().GetValue()
		s2 := o2.GetSlaveId().GetValue()

		split1 := strings.Split(s1, "-S")
		split2 := strings.Split(s2, "-S")

		o1id, _ := strconv.Atoi(split1[len(split1)-1])
		o2id, _ := strconv.Atoi(split2[len(split2)-1])

		return o1id > o2id
	}
	By(byID).Sort(offers)

	// New part: Printing the contents of the offers after sorting, including memory information.
	fmt.Println("Sorted Offers:")
	for _, offer := range offers {
		// Initialize memory value to print. Assuming a default or not found case as 0 or any relevant default.
		var memory float64
		for _, resource := range offer.GetResources() {
			if resource.GetName() == "mem" { // or "memory" depending on how it's identified
				memory = resource.GetScalar().GetValue() // Assuming scalar value for memory.
				break
			}
		}

		// Print offer details including memory information.
		fmt.Printf("Offer ID: %s, Slave ID: %s, Memory: %f\n", offer.GetId().GetValue(), offer.GetSlaveId().GetValue(), memory)
	}
}

func sortByMaxMemAvailable(offers []*mesosproto.Offer) {
    byMaxMem := func(o1, o2 *mesosproto.Offer) bool {
        // Assuming GetResources returns a slice of resources,
        // and each resource has a GetName method and a GetValue method.
        var mem1, mem2 float64 // Assuming memory value could be a floating point.
        for _, resource := range o1.GetResources() {
            if resource.GetName() == "mem" { // Assuming "mem" identifies memory.
                mem1 = resource.GetValue()
                break
            }
        }
        for _, resource := range o2.GetResources() {
            if resource.GetName() == "mem" {
                mem2 = resource.GetValue()
                break
            }
        }
        return mem1 > mem2 // Change to < if you want to sort by least memory available instead.
    }
    By(byMaxMem).Sort(offers)
}
