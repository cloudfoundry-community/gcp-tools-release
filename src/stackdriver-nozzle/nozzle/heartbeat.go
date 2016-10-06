package nozzle

import (
	"time"

	"github.com/cloudfoundry/lager"
)

type Heartbeater interface {
	Start()
	AddCounter()
}

type heartbeat struct {
	logger  lager.Logger
	trigger chan time.Time
	counter chan struct{}
}

func NewHeartbeat(logger lager.Logger, trigger chan time.Time) Heartbeater {
	counter := make(chan struct{})
	return &heartbeat{
		logger:  logger,
		trigger: trigger,
		counter: counter,
	}
}

func (h *heartbeat) Start() {
	go func() {
		eventCount := 0
		for {
			select {
			case <-h.trigger:
				h.logger.Info("counter", lager.Data{
					"eventCount": eventCount,
				})
				eventCount = 0
			case <-h.counter:
				eventCount++
			}
		}
	}()
}

func (h *heartbeat) AddCounter() {
	h.counter <- struct{}{}
}
