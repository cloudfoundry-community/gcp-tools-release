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
	"expvar"
	"sync"
)

type Counter struct {
	expvar.Int
}

func (c *Counter) Increment() {
	c.Add(1)
}

type CounterMap struct {
	expvar.Map
	category string
}

func (cm *CounterMap) Category() string {
	return cm.category
}

// A MetricSet contains a map of metric prefixes and enables
// the creation of new prefixes which are automatically exported.
type MetricSet struct {
	prefixes map[string]*MetricPrefix
	mu       sync.Mutex
}

var metricSet = &MetricSet{
	prefixes: map[string]*MetricPrefix{},
}

// NewMetricPrefix creates a new MetricPrefix which prepends "prefix/"
// to all metrics created and exported by that prefix.
func (ms *MetricSet) NewMetricPrefix(prefix string) *MetricPrefix {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if mp, ok := ms.prefixes[prefix]; ok {
		return mp
	}
	ms.prefixes[prefix] = &MetricPrefix{
		prefix:   prefix + "/",
		counters: []string{},
	}
	return ms.prefixes[prefix]
}

// Do runs f across all variables exported by all prefixes.
// The prefix map is locked during the iteration,
// but existing entries may be concurrently updated.
func (ms *MetricSet) Do(f func(expvar.KeyValue)) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, mp := range ms.prefixes {
		mp.Do(f)
	}
}

// Do runs f across all variables exported by all prefixes.
// The default prefix map is locked during the iteration,
// but existing entries may be concurrently updated.
func Do(f func(expvar.KeyValue)) {
	metricSet.Do(f)
}

// A MetricPrefix contains a list of counter names that have been
// exported via package expvar, and allows them to be manipulated.
type MetricPrefix struct {
	prefix   string
	counters []string
	mu       sync.Mutex
}

var nozzleMetrics = NewMetricPrefix("stackdriver-nozzle")

// NewMetricPrefix creates a new MetricPrefix in the default metric set.
func NewMetricPrefix(prefix string) *MetricPrefix {
	return metricSet.NewMetricPrefix(prefix)
}

// NewCounter creates and exports a new Counter for the MetricPrefix.
func (mp *MetricPrefix) NewCounter(name string) *Counter {
	v := new(Counter)
	mp.publish(name, v)
	return v
}

// NewCounter creates and exports a new Counter with the prefix "stackdriver-nozzle".
func NewCounter(name string) *Counter {
	return nozzleMetrics.NewCounter(name)
}

// NewCounterMap creates and exports a new CounterMap for the MetricPrefix.
func (mp *MetricPrefix) NewCounterMap(name, category string) *CounterMap {
	v := new(CounterMap)
	v.category = category

	mp.publish(name, v)
	return v
}

// NewCounter creates and exports a new CounterMap with the prefix "stackdriver-nozzle".
func NewCounterMap(name, category string) *CounterMap {
	return nozzleMetrics.NewCounterMap(name, category)
}

func (mp *MetricPrefix) publish(name string, v expvar.Var) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.counters = append(mp.counters, name)
	expvar.Publish(mp.prefix+name, v)
}

// Do calls f for each exported variable.
// The global counter list is locked during the iteration,
// but existing entries may be concurrently updated.
func (mp *MetricPrefix) Do(f func(expvar.KeyValue)) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, k := range mp.counters {
		val := Get(k)
		f(expvar.KeyValue{Key: k, Value: val.(expvar.Var)})
	}
}

// Get retrieves a named exported variable from the MetricPrefix.
// It returns nil if the name has not been registered.
func (mp *MetricPrefix) Get(name string) expvar.Var {
	return expvar.Get(mp.prefix + name)
}

// Get retrieves a named exported variable from the "stackdriver-nozzle"
// prefix. It returns nil if the name has not been registered.
func Get(name string) expvar.Var {
	return nozzleMetrics.Get(name)
}
