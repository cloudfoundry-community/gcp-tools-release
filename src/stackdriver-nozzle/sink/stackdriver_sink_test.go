package sink_test

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"

	. "github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/sink"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sink", func() {

	var (
		mockStackdriverClient *MockStackdriverClient
		sink                  logging.Logging
	)

	BeforeEach(func() {
		mockStackdriverClient = &MockStackdriverClient{}
		sink = NewStackdriverSink(mockStackdriverClient)
	})

	It("sink posts to client", func() {
		posted := false
		mockStackdriverClient.PostFn = func(interface{}) {
			posted = true
		}

		sink.ShipEvents(map[string]interface{} {
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
