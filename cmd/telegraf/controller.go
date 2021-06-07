package main

import (
	cbUtility "github.com/influxdata/telegraf/cloudbarista/utility"
	"os"
)

type PushController struct {
	pushControllerChan chan bool
	pushModuleChan     chan bool
	pushModuleStopChan chan bool
	pushCheckChan      chan bool
	signals            chan os.Signal
	inputFilters       []string
	outputFilters      []string
	aggregatorFilters  []string
	processorFilters   []string
	isPushOn           bool
}

func NewAgentPushController(pushControllerChan, pushModuleChan, pushCheckChan chan bool, signals chan os.Signal, inputFilters, outputFilters, aggregatorFilters, processorFilters []string) PushController {
	pushModuleStopChan := make(chan bool, 1)
	return PushController{
		pushControllerChan: pushControllerChan,
		pushModuleChan:     pushModuleChan,
		pushModuleStopChan: pushModuleStopChan,
		pushCheckChan:      pushCheckChan,
		signals:            signals,
		inputFilters:       inputFilters,
		outputFilters:      outputFilters,
		aggregatorFilters:  aggregatorFilters,
		processorFilters:   processorFilters,
		isPushOn:           cbUtility.OFF,
	}
}

func (pushController *PushController) StartPushController() {
	for {
		select {
		// API 기반 PUSH ON/OFF 제어 채널
		case pushModuleTrigger := <-pushController.pushModuleChan:
			if pushModuleTrigger {
				go func() {
					if !pushController.isPushOn {
						pushController.pushCheckChan <- cbUtility.ON
					}
					pushController.isPushOn = cbUtility.ON
					pushController.run()
				}()
			} else {
				pushController.pushModuleStopChan <- cbUtility.OFF
			}
		// OS Interrupt 신호 관련 처리
		case controllerTrigger := <-pushController.pushControllerChan:
			if !controllerTrigger {
				return
			}
		}
	}
}
