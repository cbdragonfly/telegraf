// +build !windows

package push

func run(inputFilters, outputFilters, aggregatorFilters, processorFilters []string, pushModuleChan chan bool) {
	stop = make(chan struct{})
	reloadLoop(
		inputFilters,
		outputFilters,
		aggregatorFilters,
		processorFilters,
		pushModuleChan,
	)
}
