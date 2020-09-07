package cbagent

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"log"
	"strings"

	//"github.com/influxdata/telegraf"
	cbagent "github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/logger"
)

// Agent 동작
func RunAgent(ctx context.Context,
	inputFilters []string,
	outputFilters []string,
) (map[string]telegraf.Metric, error) {
	log.Println("===============================================================================================================================")
	log.Printf("I! Starting Cloud-Barista Ondemand Monitoring Telegraf Agent")

	// If no other options are specified, load the config file and run.
	c := config.NewConfig()
	c.InputFilters = inputFilters
	c.OutputFilters = outputFilters
	err := c.LoadConfig(*usage.FConfig)
	if err != nil {
		return nil, err
	}

	if *usage.FConfigDirectory != "" {
		err = c.LoadDirectory(*usage.FConfigDirectory)
		if err != nil {
			return nil, err
		}
	}

	if *usage.FPlugins == "" && len(c.Inputs) == 0 {
		return nil, errors.New("Error: no inputs found, did you provide a valid config file?")
	}

	if int64(c.Agent.Interval.Duration) <= 0 {
		return nil, fmt.Errorf("Agent interval must be positive, found %s",
			c.Agent.Interval.Duration)
	}

	if int64(c.Agent.FlushInterval.Duration) <= 0 {
		return nil, fmt.Errorf("Agent flush_interval must be positive; found %s",
			c.Agent.Interval.Duration)
	}

	cbag, err := cbagent.NewAgent(c)
	if err != nil {
		return nil, err
	}

	// Setup logging as configured.
	logConfig := logger.LogConfig{
		Logfile: cbag.Config.Agent.Logfile,
	}

	logger.SetupLogging(logConfig)

	log.Printf("I! Loaded inputs: %s", strings.Join(c.InputNames(), " "))
	log.Printf("I! Tags enabled: %s", c.ListTags())

	return cbag.Ondemand(ctx, c.Agent.Interval.Duration)
}
