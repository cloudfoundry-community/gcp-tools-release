package mocks

func NewHeartbeater() *Heartbeater {
	return &Heartbeater{false, map[string]int{}}
}

type Heartbeater struct {
	Started  bool
	Counters map[string]int
}

func (h *Heartbeater) Start() {
	h.Started = true
}

func (h *Heartbeater) Increment(name string) {
	h.Counters[name] += 1
}

func (h *Heartbeater) Stop() {
	h.Started = false
}
