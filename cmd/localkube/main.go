package main

import (
	"fmt"
	"os"
	"os/signal"

	"rsprd.com/localkube"
)

var (
	LK *localkube.LocalKube

	DNSDomain = "cluster.local"
)

func init() {
	if name := os.Getenv("DNS_DOMAIN"); len(name) != 0 {
		DNSDomain = name
	}
}

func load() {
	LK = new(localkube.LocalKube)

	// setup etc
	etcd, err := localkube.NewEtcd(localkube.KubeEtcdClientURLs, localkube.KubeEtcdPeerURLs, "kubeetcd", localkube.KubeEtcdDataDirectory)
	if err != nil {
		panic(err)
	}
	LK.Add(etcd)

	// setup apiserver
	apiserver := localkube.NewAPIServer()
	LK.Add(apiserver)

	// setup controller-manager
	controllerManager := localkube.NewControllerManagerServer()
	LK.Add(controllerManager)

	// setup scheduler
	scheduler := localkube.NewSchedulerServer()
	LK.Add(scheduler)

	// setup kubelet (configured for weave proxy)
	kubelet := localkube.NewKubeletServer()
	LK.Add(kubelet)

	// proxy
	proxy := localkube.NewProxyServer()
	LK.Add(proxy)

	dns, err := localkube.NewDNSServer(DNSDomain, localkube.APIServerURL)
	if err != nil {
		panic(err)
	}
	LK.Add(dns)
}

func main() {
	// check for network

	// if first
	load()
	err := LK.Run(os.Args, os.Stderr)
	if err != nil {
		fmt.Printf("localkube errored: %v\n", err)
		os.Exit(1)
	}
	defer LK.StopAll()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	<-interruptChan
	fmt.Printf("\nShutting down...\n")
}
