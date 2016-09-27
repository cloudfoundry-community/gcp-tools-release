package sink

/*

type Logging interface {
	Connect() bool
	ShipEvents(map[string]interface{}, string)
}



*/

type stackdriverSink struct {
	client StackdriverClient
}

func NewStackdriverSink(client StackdriverClient) *stackdriverSink {
	return &stackdriverSink{
		client: client,
	}
}

func (s *stackdriverSink) Connect() bool {
	return false
}

func (s *stackdriverSink) ShipEvents(event map[string]interface{}, _ string) {
	s.client.Post(event)
}
