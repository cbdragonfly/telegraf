// +build !windows

package main

func (pushController *PushController) run(inputFilters, outputFilters, aggregatorFilters, processorFilters []string) {
	//stop = make(chan struct{})
	pushController.reloadLoop(
		inputFilters,
		outputFilters,
		aggregatorFilters,
		processorFilters,
		//pushController.pushModuleChan,
		//pushController.signals,
	)
}
