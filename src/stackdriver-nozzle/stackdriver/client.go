package stackdriver

import (
	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
)

type Client interface {
	Post(payload interface{})
}

type client struct {
	logger *logging.Logger
}

// TODO Auth? We're currently relying on auto-auth
// TODO error handling
func NewClient(projectID string) Client {
	ctx := context.Background()

	loggingClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	loggingClient.OnError = func(err error) {
		panic(err)
	}

	logger := loggingClient.Logger("logs/cf_logs")

	return &client{logger: logger}
}

func (s *client) Post(payload interface{}) {
	entry := logging.Entry{
		Payload: payload,
	}
	s.logger.Log(entry)
	s.logger.Flush() // TODO ???
}