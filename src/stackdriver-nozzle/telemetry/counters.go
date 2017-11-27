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

// A MetricPrefix is a path element prepended to metric names.
type MetricPrefix string

// The metricSet contains a map of metric prefixes and enables
// the creation of new prefixes which are automatically exported.
var metricSet = struct {
	prefixes map[MetricPrefix][]string
	mu       sync.Mutex
}{
	prefixes: map[MetricPrefix][]string{},
}

// stackdriverNozzle is the prefix where the package-level telemetry
// functions place their counters.
var stackdriverNozzle MetricPrefix = "stackdriver-nozzle"

// NewCounter creates and exports a new Counter for the MetricPrefix.
func (mp MetricPrefix) NewCounter(name string) *Counter {
	v := new(Counter)
	mp.publish(name, v)
	return v
}

// NewCounter creates and exports a new Counter with the prefix "stackdriver-nozzle".
func NewCounter(name string) *Counter {
	return stackdriverNozzle.NewCounter(name)
}

// NewCounterMap creates and exports a new CounterMap for the MetricPrefix.
func (mp MetricPrefix) NewCounterMap(name, category string) *CounterMap {
	v := new(CounterMap)
	v.category = category

	mp.publish(name, v)
	return v
}

// NewCounter creates and exports a new CounterMap with the prefix "stackdriver-nozzle".
func NewCounterMap(name, category string) *CounterMap {
	return stackdriverNozzle.NewCounterMap(name, category)
}

func (mp MetricPrefix) qualify(name string) string {
	return string(mp) + "/" + name
}

func (mp MetricPrefix) publish(name string, v expvar.Var) {
	metricSet.mu.Lock()
	defer metricSet.mu.Unlock()
	if _, ok := metricSet.prefixes[mp]; !ok {
		metricSet.prefixes[mp] = []string{name}
	} else {
		metricSet.prefixes[mp] = append(metricSet.prefixes[mp], name)
	}
	expvar.Publish(mp.qualify(name), v)
}

// Do calls f for each exported variable.
// The global metric set is locked during the iteration,
// but existing entries may be concurrently updated.
func Do(f func(expvar.KeyValue)) {
	metricSet.mu.Lock()
	defer metricSet.mu.Unlock()
	for _, mp := range metricSet.prefixes {
		for _, k := range mp {
			val := Get(k)
			f(expvar.KeyValue{Key: k, Value: val.(expvar.Var)})
		}
	}
}

// Get retrieves a named exported variable with the given MetricPrefix.
// It returns nil if the name has not been registered.
func (mp MetricPrefix) Get(name string) expvar.Var {
	return expvar.Get(mp.qualify(name))
}

// Get retrieves a named exported variable from the "stackdriver-nozzle"
// prefix. It returns nil if the name has not been registered.
func Get(name string) expvar.Var {
	return stackdriverNozzle.Get(name)
}
