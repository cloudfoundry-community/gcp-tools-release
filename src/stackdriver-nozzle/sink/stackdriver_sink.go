package sink

import "github.com/cloudfoundry-community/firehose-to-syslog/logging"

type stackdriverSink struct {
	client StackdriverClient
}

func NewStackdriverSink(client StackdriverClient) logging.Logging {
	return &stackdriverSink{
		client: client,
	}
}

func (s *stackdriverSink) Connect() bool {
	return false
}

func (s *stackdriverSink) ShipEvents(event map[string]interface{}, _ string /* TODO research second string */) {
	s.client.Post(event)
}
