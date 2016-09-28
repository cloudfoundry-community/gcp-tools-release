package firehose

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehoseclient"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
)

type Client interface {
	StartListening(nozzle logging.Logging) error
}

type client struct {
	config *ClientConfig
}

type ClientConfig struct {
	User string
	Password string
	ApiEndpoint string
	SkipSSLValidation bool
}

func NewClient(config *ClientConfig) (Client) {
	return &client{config:config}
}

func (f *client) StartListening(nozzle logging.Logging) error {
	config := f.config

	cfConfig := &cfclient.Config{
		ApiAddress:        config.ApiEndpoint,
		Username:          config.User,
		Password:          config.Password,
		SkipSslValidation: config.SkipSSLValidation,
	}

	cfClient := cfclient.NewClient(cfConfig)

	cachingClient := caching.NewCachingEmpty()

	eventRouter := eventRouting.NewEventRouting(cachingClient, nozzle)

	//err := eventRouter.SetupEventRouting("LogMessage,ValueMetric,HttpStartStop,CounterEvent,Error,ContainerMetric")
	err := eventRouter.SetupEventRouting("HttpStartStop")
	if err != nil {
		return err
	}

	if nozzle.Connect() {
		firehoseConfig := &firehoseclient.FirehoseConfig{
			TrafficControllerURL:   cfClient.Endpoint.DopplerEndpoint,
			InsecureSSLSkipVerify:  true,
			IdleTimeoutSeconds:     30,
			FirehoseSubscriptionID: "stackdriver-nozzle",
		}

		firehoseClient := firehoseclient.NewFirehoseNozzle(cfClient, eventRouter, firehoseConfig)

		firehoseClient.Start()
	}
	return nil
}
