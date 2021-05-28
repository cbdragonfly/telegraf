package main

import (
	"os"
	"sync"
)

type PushController struct {
	pushControllerChan chan bool
	pushModuleChan     chan bool
	pushModuleStopChan chan bool
	signals            chan os.Signal
	wg                 sync.WaitGroup
	inputFilters       []string
	outputFilters      []string
	aggregatorFilters  []string
	processorFilters   []string
}

func NewAgentPushController(pushControllerChan chan bool, pushModuleChan chan bool, signals chan os.Signal, inputFilters, outputFilters, aggregatorFilters, processorFilters []string) PushController {
	pushModuleStopChan := make(chan bool, 1)
	return PushController{
		pushControllerChan: pushControllerChan,
		pushModuleChan:     pushModuleChan,
		pushModuleStopChan: pushModuleStopChan,
		signals:            signals,
		inputFilters:       inputFilters,
		outputFilters:      outputFilters,
		aggregatorFilters:  aggregatorFilters,
		processorFilters:   processorFilters,
	}
}

func (pushController *PushController) StartPushController() {
	for {
		select {
		case controllerTrigger := <-pushController.pushControllerChan:
			if !controllerTrigger {
				pushController.pushModuleChan <- false
			} else {
				pushController.pushModuleChan <- true
				go func() {
					for {
						select {
						case pushModuleTrigger := <-pushController.pushModuleChan:
							if pushModuleTrigger {
								go pushController.run()
							} else {
								pushController.pushModuleStopChan <- false
								return
							}
						}
					}
				}()
			}
		}
	}
}
