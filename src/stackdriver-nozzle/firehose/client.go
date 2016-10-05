package firehose

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

type FirehoseHandler interface {
	HandleEvent(*events.Envelope) error
}

type Client interface {
	StartListening(FirehoseHandler) error
}

type client struct {
	cfConfig *cfclient.Config
}

func NewClient(apiAddress, username, password string, skipSSLValidation bool) Client {
	return &client{
		&cfclient.Config{
			ApiAddress:        apiAddress,
			Username:          username,
			Password:          password,
			SkipSslValidation: skipSSLValidation}}
}

func (c *client) StartListening(fh FirehoseHandler) error {
	cfConfig := &cfclient.Config{
		ApiAddress:        c.cfConfig.ApiAddress,
		Username:          c.cfConfig.Username,
		Password:          c.cfConfig.Password,
		SkipSslValidation: c.cfConfig.SkipSslValidation}
	cfClient := cfclient.NewClient(cfConfig)

	// PRECHECKIN: remove
	endpoint := "wss://doppler.104.199.124.149.xip.io:443"

	fmt.Println("-----")
	fmt.Println(cfClient.Endpoint.DopplerEndpoint)
	fmt.Println("-----")

	cfConsumer := consumer.New(
		endpoint,
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
