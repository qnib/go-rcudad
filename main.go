package main

import (
	"bytes"
	"fmt"
		"log"
	"os/exec"
		"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
	"bufio"
	"github.com/qnib/go-rcudad/prom"
	"os"
		"github.com/codegangsta/cli"
)

func startDaemon(m prom.Metrics) {
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
			m.CheckLine(line)
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderrIn)
		for scanner.Scan() {
			line := scanner.Text()
			m.CheckLine(line)
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

func startProm(ctx *cli.Context) {
	addr := ctx.String("listen-addr")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))

}

func Run(ctx *cli.Context) {
    // start prometheus
	go startProm(ctx)
	m := prom.NewMetrics(ctx.Bool("debug"))
	m.AddRMatchCnt(
		"app_finished", "daemon", "How often was the rCUDA client finished correctly.",
		`.*: Remote application finished with \d+ threads.`)
	m.AddRMatchCnt(
		"device_1_set", "daemon", "Device 1 set for the first time.",
		`.*: Device 1 set for first time.`)
	m.AddRMatchCnt(
		"device_0_init", "daemon", "CUDA initialized on device 0.",
		`CUDA initialized on device 0.`)
	m.Register()
	for {
		startDaemon(m)
		m.CounterInc("restart_count")
		time.Sleep(1*time.Second)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Daemon to fire up rCUDAd."
	app.Usage = "go-rcudad [options]"
	app.Version = "0.1.2"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen-addr",
			Value: "0.0.0.0:8081",
			Usage: "IP:PORT to bind prometheus",
			EnvVar: "RCUDAD_PROMETHEUS_ADDR",
		},
		cli.BoolFlag{
			Name: "debug",
			Usage: "Be more verbose..",
			EnvVar: "RCUDAD_DEBUG",

		},
	}
	app.Action = Run
	app.Run(os.Args)
}
