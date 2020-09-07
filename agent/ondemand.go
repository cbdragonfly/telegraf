package agent

import (
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf/models"
)

func (a *Agent) ondemandrunOutputs(
	unit *outputUnit,
) map[string]telegraf.Metric {
	var wg sync.WaitGroup

	// Start flush loop
	interval := a.Config.Agent.FlushInterval.Duration
	jitter := a.Config.Agent.FlushJitter.Duration

	ctx, cancel := context.WithCancel(context.Background())

	for _, output := range unit.outputs {
		interval := interval
		// Overwrite agent flush_interval if this plugin has its own.
		if output.Config.FlushInterval != 0 {
			interval = output.Config.FlushInterval
		}

		jitter := jitter
		// Overwrite agent flush_jitter if this plugin has its own.
		if output.Config.FlushJitter != 0 {
			jitter = output.Config.FlushJitter
		}

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			ticker := NewRollingTicker(interval, jitter)
			defer ticker.Stop()

			a.flushLoop(ctx, output, ticker)
		}(output)
	}

	cancel()
	//var Datalist = []telegraf.Metric{}
	var Datalist = map[string]telegraf.Metric{}
	for metric := range unit.src {
		Datalist[metric.Name()] = metric
	}
	wg.Wait()
	return Datalist
}

func (a *Agent) Ondemand(ctx context.Context, wait time.Duration) (map[string]telegraf.Metric, error) {
	data, err := a.ondemand(ctx, wait)
	if err != nil {
		return nil, err
	}

	if models.GlobalGatherErrors.Get() != 0 {
		return nil, fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
	}

	unsent := 0
	for _, output := range a.Config.Outputs {
		unsent += output.BufferLength()
	}
	if unsent != 0 {
		return nil, fmt.Errorf("output plugins unable to send %d metrics", unsent)
	}

	return data, nil
}

func (a *Agent) ondemand(ctx context.Context, wait time.Duration) (map[string]telegraf.Metric, error) {
	log.Printf("D! [agent] Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return nil, err
	}

	log.Printf("D! [agent] Connecting outputs")
	next, ou, err := a.startOutputs(ctx, a.Config.Outputs)
	if err != nil {
		return nil, err
	}

	/*
		var apu []*processorUnit
		var au *aggregatorUnit
		if len(a.Config.Aggregators) != 0 {
			procC := next
			if len(a.Config.AggProcessors) != 0 {
				procC, apu, err = a.startProcessors(next, a.Config.AggProcessors)
				if err != nil {
					return nil, err
				}
			}

			next, au, err = a.startAggregators(procC, next, a.Config.Aggregators)
			if err != nil {
				return nil, err
			}
		}

		var pu []*processorUnit
		if len(a.Config.Processors) != 0 {
			next, pu, err = a.startProcessors(next, a.Config.Processors)
			if err != nil {
				return nil, err
			}
		}
	*/
	iu, err := a.testStartInputs(next, a.Config.Inputs)
	if err != nil {
		return nil, err
	}
	log.Println("Start Gathering Metric")
	var data = map[string]telegraf.Metric{}
	cbchannel := make(chan map[string]telegraf.Metric)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go func() {
			cbchannel <- a.ondemandrunOutputs(ou)
		}()

	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		data = <-cbchannel
	}()

	/*
		if au != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = a.runProcessors(apu)
				if err != nil {
					log.Printf("E! [agent] Error running processors: %v", err)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				err = a.runAggregators(startTime, au)
				if err != nil {
					log.Printf("E! [agent] Error running aggregators: %v", err)
				}
			}()
		}

		if pu != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = a.runProcessors(pu)
				if err != nil {
					log.Printf("E! [agent] Error running processors: %v", err)
				}
			}()
		}*/

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = a.testRunInputs(ctx, wait, iu)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()
	log.Printf("D! [agent] Stopped Successfully")
	log.Println("Completed Gathering Metric")
	return data, nil
}
