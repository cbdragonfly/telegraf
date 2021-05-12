package main

import (
	cbapiserver "github.com/influxdata/telegraf/cloudbarista/apiserver"
	"github.com/influxdata/telegraf/cloudbarista/push"
	"log"
	"sync"
)

var wg sync.WaitGroup

func main() {
	pushControllerChan := make(chan bool, 1)
	pushModuleChan := make(chan bool, 1)

	wg.Add(1)
	go func() {
		cbapiServer := cbapiserver.NewAgentAPIServer(8888, pushControllerChan)
		err := cbapiServer.RunAPIServer()
		if err != nil {
			log.Fatalf("Running Cloud-Barista API Server Error")
		}
	}()

	wg.Add(1)
	go func() {
		push.StartPushController(pushControllerChan, pushModuleChan)
	}()
	wg.Wait()
	return
}
