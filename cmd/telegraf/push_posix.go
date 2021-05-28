// +build !windows

package main

func (pushController *PushController) run() {
	//stop = make(chan struct{})
	pushController.reloadLoop(
		pushController.inputFilters,
		pushController.outputFilters,
		pushController.aggregatorFilters,
		pushController.processorFilters,
	)
}
