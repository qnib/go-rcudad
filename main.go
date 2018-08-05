package main

import (
	"bytes"
	"fmt"
		"log"
	"os/exec"
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
	"bufio"
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
	m.gauges["log_count"] = prometheus.NewGauge(prometheus.GaugeOpts{
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

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")


func startDaemon(m Metrics) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("./rCUDAd", "-iv")

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdoutIn)
		for scanner.Scan() {
			line := scanner.Text()
			// Log the stdout
			m.gauges["log_count"].Inc()
			fmt.Printf("stdin>> %s\n", line)
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderrIn)
		for scanner.Scan() {
			line := scanner.Text()
			// Log the stderr
			m.gauges["log_count"].Inc()
			fmt.Printf("stderr>> %s\n", line)
		}
	}()

	err = cmd.Wait()
	if err != nil {
		// Count error code
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

func startProm() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))

}

func main() {
    // start prometheus
	go startProm()
	m := NewMetrics()
	m.Register()
	for {
		startDaemon(m)
		m.counter["restart_count"].Inc()
		time.Sleep(1*time.Second)
	}
}
