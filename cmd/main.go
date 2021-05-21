package main

import (
	"context"
	"github.com/influxdata/telegraf/cloudbarista/listener"
	"github.com/influxdata/telegraf/cloudbarista/push"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

func main() {
	pushControllerChan := make(chan bool, 1)
	pushModuleChan := make(chan bool, 1)
	signals := make(chan os.Signal, 100)

	ctx, cancel := context.WithCancel(context.Background())
	CB_Listener := listener.NewAgentPullListener(8888, pushControllerChan, signals, ctx, cancel)

	wg.Add(1)
	go func() {
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGINT)
		<-signals
		<-signals
		wg.Done()
		return
	}()

	CB_Push_Controller := push.NewAgentPushController(pushControllerChan, pushModuleChan, signals)
	wg.Add(1)
	go func() {
		err := CB_Listener.Start()
		if err != nil {
			log.Fatalf("Running Cloud-Barista API Server Error")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		CB_Push_Controller.StartPushController()
		//push.StartPushController(pushControllerChan, pushModuleChan, signals)
		wg.Done()
	}()
	wg.Wait()
	return
}
