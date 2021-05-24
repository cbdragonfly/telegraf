package main

import (
	"os"
)

type PushController struct {
	pushControllerChan chan bool
	pushModuleChan     chan bool
	signals            chan os.Signal
}

func NewAgentPushController(pushControllerChan chan bool, pushModuleChan chan bool, signals chan os.Signal) PushController {
	return PushController{
		pushControllerChan: pushControllerChan,
		pushModuleChan:     pushModuleChan,
		signals:            signals,
	}
}

func (pushController *PushController) StartPushController() {
	for {
		select {
		case controllerTrigger := <-pushController.pushControllerChan:
			switch controllerTrigger {
			case false:
				pushController.pushModuleChan <- false
				break
			case true:
				pushController.pushModuleChan <- true
				go func() {
					select {
					case pushModuleTrigger := <-pushController.pushModuleChan:
						if pushModuleTrigger {
							pushController.startPushMonitoring()
						}
					}
				}()
				break
			}
		}
	}
}
