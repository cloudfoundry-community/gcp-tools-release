package heartbeat

import (
	"errors"
	"time"

	"github.com/cloudfoundry/lager"
)

type Heartbeater interface {
	Start()
	Increment(string)
	Stop()
}

type heartbeat struct {
	logger  lager.Logger
	trigger <-chan time.Time
	counter chan string
	done    chan struct{}
	started bool
}

func NewHeartbeat(logger lager.Logger, trigger <-chan time.Time) Heartbeater {
	counter := make(chan string)
	done := make(chan struct{})
	return &heartbeat{
		logger:  logger,
		trigger: trigger,
		counter: counter,
		done:    done,
		started: false,
	}
}

func (h *heartbeat) Start() {
	h.started = true
	go func() {
		counters := map[string]uint{}
		for {
			select {
			case <-h.trigger:
				h.logger.Info(
					"heartbeat", lager.Data{"counters": counters},
				)
				counters = map[string]uint{}
			case name := <-h.counter:
				counters[name] += 1
			case <-h.done:
				h.logger.Info(
					"heartbeat", lager.Data{"counters": counters},
				)
				return
			}
		}
	}()
}

func (h *heartbeat) Increment(name string) {
	if h.started {
		h.counter <- name
	} else {
		h.logger.Error(
			"heartbeat",
			errors.New("attempted to increment counter without starting heartbeat"),
		)
	}
}

func (h *heartbeat) Stop() {
	h.done <- struct{}{}
	h.started = false
}
