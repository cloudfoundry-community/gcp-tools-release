package firehose

import (
	"crypto/tls"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/lager"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

type FirehoseHandler interface {
	HandleEvent(*events.Envelope) error
}

type Client interface {
	StartListening(FirehoseHandler) error
	EnsureCfClient() *cfclient.Client
}

type client struct {
	cfConfig *cfclient.Config
	logger   lager.Logger
	cfClient *cfclient.Client
}

func NewClient(apiAddress, username, password string, skipSSLValidation bool, logger lager.Logger) Client {
	return &client{
		logger: logger,
		cfConfig: &cfclient.Config{
			ApiAddress:        apiAddress,
			Username:          username,
			Password:          password,
			SkipSslValidation: skipSSLValidation,
		},
		cfClient: nil,
	}
}

func (c *client) EnsureCfClient() *cfclient.Client {
	if c.cfClient == nil {
		c.cfClient = cfclient.NewClient(c.cfConfig)
	}

	return c.cfClient
}

func (c *client) StartListening(fh FirehoseHandler) error {
	cfClient := c.EnsureCfClient()
	cfConsumer := consumer.New(
		cfClient.Endpoint.DopplerEndpoint,
		&tls.Config{InsecureSkipVerify: c.cfConfig.SkipSslValidation},
		nil)

	refresher := CfClientTokenRefresh{cfClient: cfClient}
	cfConsumer.SetIdleTimeout(time.Duration(30) * time.Second)
	cfConsumer.RefreshTokenFrom(&refresher)
	messages, errs := cfConsumer.FirehoseWithoutReconnect("test", "")

	for {
		select {
		case envelope := <-messages:
			err := fh.HandleEvent(envelope)
			if err != nil {
				return err
			}
		case err := <-errs:
			return err
		}
	}
}

type CfClientTokenRefresh struct {
	cfClient *cfclient.Client
}

func (ct *CfClientTokenRefresh) RefreshAuthToken() (string, error) {
	return ct.cfClient.GetToken(), nil
}
