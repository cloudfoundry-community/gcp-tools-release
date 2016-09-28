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

	It("ships events to the stackdriver client", func() {
		var postedEvent interface{}
		mockStackdriverClient.PostFn = func(e interface{}) {
			postedEvent = e
		}

		shippedEvent := map[string]interface{}{
			"event_type": "HttpStartStop",
			"foo": "bar",
		}

		n := nozzle.Nozzle{StackdriverClient: mockStackdriverClient}
		n.ShipEvents(shippedEvent, "message")

		Expect(postedEvent).To(Equal(shippedEvent))
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
