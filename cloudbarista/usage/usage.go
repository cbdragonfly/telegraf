package usage

import (
	"context"
	"flag"
	"github.com/influxdata/telegraf"
)

//JSON 응답을 위한 선언
type CBMetricType struct {
	Name      string
	Tags      map[string]string
	Fields    map[string]interface{}
	Timestamp int64
}

var (
	FConfig          = flag.String("Config", "", "configuration file to load")
	FConfigDirectory = flag.String("Config-directory", "",
		"directory containing additional *.conf files")
	FPlugins = flag.String("Plugin-directory", "",
		"path to directory containing external plugins")
	Ctx           context.Context
	InputFilters  []string
	OutputFilters []string
)

func ExtractMetric(target string, insert map[string]CBMetricType) map[string]CBMetricType {
	var result = map[string]CBMetricType{}
	for key, _ := range insert {
		if insert[key].Name == target {
			result[target] = CBMetricType{
				Name:      insert[key].Name,
				Tags:      insert[key].Tags,
				Fields:    insert[key].Fields,
				Timestamp: insert[key].Timestamp,
			}
		}
	}
	return result
}
func ToCBMetric(insert map[string]telegraf.Metric) map[string]CBMetricType {
	var result = map[string]CBMetricType{}
	for key, _ := range insert {
		result[insert[key].Name()] = CBMetricType{
			Name:      insert[key].Name(),
			Tags:      insert[key].Tags(),
			Fields:    insert[key].Fields(),
			Timestamp: insert[key].Time().Unix(),
		}
	}
	return result
}
