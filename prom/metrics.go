package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"fmt"
)

type Metrics struct {
	gauges map[string]prometheus.Gauge
	counter map[string]prometheus.Counter
}


func NewMetrics() Metrics {
	m := Metrics{
		gauges: map[string]prometheus.Gauge{},
		counter: map[string]prometheus.Counter{},
	}
	m.counter["log_count"] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "rcuda",
		Subsystem: "daemon",
		Name:      "log_count",
		Help:      "Number of loglines",
	})
	m.counter["restart_count"] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "rcuda",
		Subsystem: "daemon",
		Name:      "restart_count",
		Help:      "How often did the loop restart",
	})
	return m
}

func (m *Metrics) CounterInc(name string) (err error) {
	if c, ok := m.counter[name]; ok {
		c.Inc()
	} else {
		err = fmt.Errorf("Could not find '%s' in Counters", name)
	}
	return err
}

func (m *Metrics) Register() {
	for k, g := range m.gauges {
		fmt.Printf("Register Prometheus Gauge: %s\n", k)
		prometheus.MustRegister(g)
	}
	for k, c := range m.counter {
		fmt.Printf("Register Prometheus Counter: %s\n", k)
		prometheus.MustRegister(c)
	}
}
