/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package telemetry

import (
	"bytes"
	"expvar"
	"fmt"
	"sort"
	"sync"
)

// A Counter is a monotonically-increasing integer expvar associated with a
// set of labels expressed as a key-value string map.
//
// Note that exporting custom labels for Counters to Stackdriver is only
// implemented for counters that are part of a CounterMap.
type Counter struct {
	expvar.Int
	Labels map[string]string
}

// Increment adds one to the counter's expvar value.
func (c *Counter) Increment() {
	c.Add(1)
}

// IntValue returns the counter's value as an int rather than an int64. Tests are
// generally written with int types, so this is useful to avoid scattering type
// casts around the test codebase.
func (c *Counter) IntValue() int {
	return int(c.Value())
}

// A CounterMap is used to export a set of related Counters which have the
// same label keys.
type CounterMap struct {
	expvar.Map
	LabelKeys []string
}

// Counter retrieves or creates a Counter with a given set of label values.
func (cm *CounterMap) Counter(labelValues ...string) (*Counter, error) {
	if len(labelValues) != len(cm.LabelKeys) {
		return nil, fmt.Errorf("want %d label values for map, got %d",
			len(cm.LabelKeys), len(labelValues))
	}
	v := &Counter{Labels: map[string]string{}}
	for i, k := range cm.LabelKeys {
		v.Labels[k] = labelValues[i]
	}
	key := MapKey(v.Labels)
	existing := cm.Get(key)
	if existing == nil {
		// TODO(fluffle): this might race?
		cm.Set(key, v)
		return v, nil
	}
	if v, ok := existing.(*Counter); ok {
		return v, nil
	}
	// Shouldn't reach here, it implies a non-Counter in the map.
	return nil, fmt.Errorf("found non-Counter %#v in map", existing)
}

// MustCounter is a version of Counter that panics on error, for use
// in init functions.
func (cm *CounterMap) MustCounter(labelValues ...string) *Counter {
	ctr, err := cm.Counter(labelValues...)
	if err == nil {
		return ctr
	}
	panic(err)
}

// MapKey serializes the set of label keys and values for use as
// the key in an expvar.Map.
// TODO(fluffle): Dedupe this with messages.Metric.Hash()?
func MapKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := &bytes.Buffer{}
	for i, k := range keys {
		buf.WriteString(fmt.Sprintf("%q=%q", k, labels[k]))
		if i+1 < len(keys) {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

// The metricSet contains a map of metric prefixes and enables
// the creation of new prefixes which are automatically exported.
var metricSet = struct {
	prefixes map[MetricPrefix][]string
	mu       sync.Mutex
}{
	prefixes: map[MetricPrefix][]string{},
}

// A MetricPrefix is a path element prepended to metric names.
type MetricPrefix string

// Nozzle is the prefix under which the nozzle exports metrics about
// its own operation. It's created here because metrics will be created
// in many places throughout the Nozzle's code base.
const Nozzle MetricPrefix = "stackdriver-nozzle"

// Qualify returns the metric name prepended by the metric prefix and "/".
func (mp MetricPrefix) Qualify(name string) string {
	return string(mp) + "/" + name
}

// NewCounter creates and exports a new Counter for the MetricPrefix.
func NewCounter(mp MetricPrefix, name string) *Counter {
	v := new(Counter)
	publish(mp, name, v)
	return v
}

// NewCounterMap creates and exports a new CounterMap for the MetricPrefix.
func NewCounterMap(mp MetricPrefix, name string, labelKeys ...string) *CounterMap {
	v := &CounterMap{LabelKeys: labelKeys}
	publish(mp, name, v)
	return v
}

func publish(mp MetricPrefix, name string, v expvar.Var) {
	metricSet.mu.Lock()
	defer metricSet.mu.Unlock()
	if _, ok := metricSet.prefixes[mp]; !ok {
		metricSet.prefixes[mp] = []string{name}
	} else {
		metricSet.prefixes[mp] = append(metricSet.prefixes[mp], name)
	}
	expvar.Publish(mp.Qualify(name), v)
}

// forEachMetric calls f for each exported variable.
// The global metric set is locked during the iteration,
// but existing entries may be concurrently updated.
func forEachMetric(f func(expvar.KeyValue)) {
	metricSet.mu.Lock()
	defer metricSet.mu.Unlock()
	for mp, counters := range metricSet.prefixes {
		for _, k := range counters {
			val := Get(mp, k)
			f(expvar.KeyValue{Key: mp.Qualify(k), Value: val.(expvar.Var)})
		}
	}
}

// Get retrieves a named exported variable with the given MetricPrefix.
// It returns nil if the name has not been registered.
func Get(mp MetricPrefix, name string) expvar.Var {
	return expvar.Get(mp.Qualify(name))
}
