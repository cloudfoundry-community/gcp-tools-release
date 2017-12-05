package messages

import (
	"bytes"
	"sort"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

// Metric represents one of the metrics contained in an events.Envelope.
type Metric struct {
	Name      string
	Labels    map[string]string `json:"-"`
	Value     float64
	EventTime time.Time
	StartTime time.Time                 `json:"-"`
	Unit      string                    // TODO Should this be "1" if it's empty?
	Type      events.Envelope_EventType `json:"-"`
}

func (m *Metric) IsCumulative() bool {
	return m.Type == events.Envelope_CounterEvent
}

func (m *Metric) Hash() string {
	var b bytes.Buffer

	// Extract keys to a slice and sort it
	numKeys := len(m.Labels) + 1
	keys := make([]string, 0, numKeys)
	keys = append(keys, m.Name)
	for k := range m.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.Write([]byte(k))
		b.Write([]byte(m.Labels[k]))
	}
	return b.String()
}
