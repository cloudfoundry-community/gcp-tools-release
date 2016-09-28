package nozzle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/nozzle"
)

var _ = Describe("Nozzle", func() {

	var (
		mockStackdriverClient *MockStackdriverClient
	)

	BeforeEach(func() {
		mockStackdriverClient = &MockStackdriverClient{}
	})

	It("sink posts to client", func() {
		posted := false
		mockStackdriverClient.PostFn = func(interface{}) {
			posted = true
		}

		n := nozzle.Nozzle { StackdriverClient: mockStackdriverClient }
		n.ShipEvents(map[string]interface{} {
			"foo": "bar",
		}, "")

		Expect(posted).To(BeTrue())
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
