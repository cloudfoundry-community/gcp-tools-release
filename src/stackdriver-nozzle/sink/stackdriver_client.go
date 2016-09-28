package sink

import (
	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
)

type StackdriverClient interface {
	Post(payload interface{})
}

type stackdriverClient struct {
	logger *logging.Logger
}

// TODO Auth? We're currently relying on auto-auth
// TODO error handling
func NewStackdriverClient() StackdriverClient {
	ctx := context.Background()

	client, err := logging.NewClient(ctx, "evandbrown17")
	if err != nil {
		panic(err)
	}

	client.OnError = func(err error) {
		panic(err)
	}

	logger := client.Logger("logs/cf_logs")

	return &stackdriverClient{logger: logger}
}

func (s *stackdriverClient) Post(payload interface{}) {
	entry := logging.Entry{
		Payload: payload,
	}
	s.logger.Log(entry)
	s.logger.Flush() // TODO ???
}