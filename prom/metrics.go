package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"fmt"
	"regexp"
)

type Metrics struct {
	debug 	bool
	gauges  map[string]prometheus.Gauge
	counter map[string]prometheus.Counter
	rMatchCnt map[string]string
}


func NewMetrics(debug bool) Metrics {
	m := Metrics{
		debug: debug,
		gauges: map[string]prometheus.Gauge{},
		counter: map[string]prometheus.Counter{},
		rMatchCnt: map[string]string{},
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

// AddRMatchCnt will increment in case the regex matches
func (m *Metrics) AddRMatchCnt(name, sub, help, reg string) {
	if _, ok := m.counter[name];ok {
		fmt.Printf("Counter '%s' already exists\n", name)
		return
	}
	m.counter[name] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "rcuda",
		Subsystem: sub,
		Name:      name,
		Help:      help,
	})
	m.rMatchCnt[name] = reg

}

func (m *Metrics) CheckLine(line string) {
	m.CounterInc("log_count")
	for k, r := range m.rMatchCnt {
		if matched, err := regexp.MatchString(r, line); err == nil && matched {
			m.CounterInc(k)
		} else if m.debug {
			fmt.Printf("    >> '%s' does not match '%s'\n", r, line)
		} else {
			fmt.Println(line)
		}
	}
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

