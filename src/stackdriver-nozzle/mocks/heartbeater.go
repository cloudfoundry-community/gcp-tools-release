package mocks

type Heartbeater struct {
	Started bool
	counter int
}

func (h *Heartbeater) Start() {
	h.Started = true
}

func (h *Heartbeater) Increment(_ string) {
	h.counter += 1
}

func (h *Heartbeater) Stop() {
	h.Started = false
}

func (h *Heartbeater) GetCounter() int {
	return h.counter
}
