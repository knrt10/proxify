package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"github.com/knrt10/proxify/pkg/config"
	"github.com/knrt10/proxify/pkg/proxy"
	"github.com/knrt10/proxify/pkg/tracer"
)

var (
	lport = 8080
)

func main() {
	ctx := newCancelableContext()

	cfgStore := config.NewConfigStore("./config.json")

	// watch for changes to the config
	ch, err := cfgStore.StartWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer cfgStore.Close()

	go func() {
		for cfg := range ch {
			fmt.Println("got config change:", cfg)
		}
	}()

	// TODO: put a proxy here :)
	var bpfEnabled bool
	// flags declaration using flag package
	flag.BoolVar(&bpfEnabled, "b", false, "enable bpf steering")
	flag.IntVar(&lport, "p", 8080, "listener port for bpf steering")
	flag.Parse()

	// Load initial config from config.json file
	var cfg config.Config
	cfg, err = cfgStore.Read()
	if err != nil {
		log.Fatalln(err)
	}

	var ports []int
	var targets []proxy.Target
	// Range through cfg to get port and target data
	for _, app := range cfg.Apps {
		for _, port := range app.Ports {
			var targets []proxy.Target
			ports = append(ports, port)
			// Create a new target instance for every target and append
			// it into a slice
			for _, target := range app.Targets {
				targets = append(targets, proxy.NewTargetServer(target))
			}
			// Create a new Loadbalancer for every port
			lb := proxy.NewLoadBalancer(port, targets)
			// Setup proxy as a different goroutine for every port
			go proxy.Setup(":"+strconv.Itoa(port), lb)
		}

		// These targets are used for port 8080 socket
		for _, target := range app.Targets {
			targets = append(targets, proxy.NewTargetServer(target))
		}
	}

	/* If ebpf steering is enabled then,
	Start the server on port 8080 for our proxify application.
	This socket pid will be used when updating BPF `proxy_sock` map.
	Server running on this port will redirect traffic to any given target
	that has the least traffic.
	*/
	if bpfEnabled {
		go listen(targets)

		// HACK: Server is now already running on port 8080. This is done
		// to get the FD of the running socket
		cmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(lport), "-FD")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}

		pidFD := strings.Split(out.String(), "\n")
		fdString := strings.Split(pidFD[1], "f")[1]
		fd, _ := strconv.Atoi(fdString)

		// Start a new eBPF tracer
		t, err := tracer.NewTracer(ports)
		if err != nil {
			log.Fatalln(err)
		}

		// Register the interface on every run
		if err := t.RegisterIface(0, fd); err != nil {
			log.Fatalln(err)
		}
	}
	<-ctx.Done()
}

// newCancelableContext returns a context that gets canceled by a SIGINT
func newCancelableContext() context.Context {
	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-doneCh
		log.Println("signal recieved")
		cancel()
	}()

	return ctx
}

// listen is used to create a TCP server on a given port
func listen(targets []proxy.Target) {
	// Create a new Loadbalancer for port 8080
	lb := proxy.NewLoadBalancer(lport, targets)
	// Setup proxy as a different goroutine for every port
	proxy.Setup(":"+strconv.Itoa(lport), lb)
}
