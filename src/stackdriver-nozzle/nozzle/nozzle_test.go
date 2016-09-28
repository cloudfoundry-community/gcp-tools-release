package nozzle_test

import (
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/nozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nozzle", func() {

	var (
		mockStackdriverClient *MockStackdriverClient
	)

	BeforeEach(func() {
		mockStackdriverClient = &MockStackdriverClient{}
	})

	It("ships something to the stackdriver client", func() {
		var postedEvent interface{}
		mockStackdriverClient.PostFn = func(e interface{}) {
			postedEvent = e
		}

		shippedEvent := map[string]interface{}{
			"event_type": "HttpStartStop",
			"foo":        "bar",
		}

		n := nozzle.Nozzle{StackdriverClient: mockStackdriverClient}
		n.ShipEvents(shippedEvent, "message")

		Expect(postedEvent).To(Equal(shippedEvent))
	})

	It("ships multiple events", func() {
		count := 0
		mockStackdriverClient.PostFn = func(e interface{}) {
			count += 1
		}

		shippedEvent := map[string]interface{}{
			"event_type": "HttpStartStop",
			"foo":        "bar",
		}
		n := nozzle.Nozzle{StackdriverClient: mockStackdriverClient}

		for i := 0; i < 10; i++ {
			n.ShipEvents(shippedEvent, "message")
		}

		Expect(count).To(Equal(10))
	})
})

type MockStackdriverClient struct {
	PostFn func(interface{})
}

func (m *MockStackdriverClient) Post(payload interface{}) {
	if m.PostFn != nil {
		m.PostFn(payload)
	}
}
