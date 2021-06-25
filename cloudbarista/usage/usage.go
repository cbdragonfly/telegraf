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

type CBMCISMetric struct {
	Result  string `json:"result"`
	Unit    string `json:"unit"`
	Desc    string `json:"desc"`
	Elapsed string `json:"elapsed"`
	SpecId  string `json:"specid"`
}

type Request struct {
	Host string `json:"host"`
	Spec string `json:"spec"`
}

type MRequest struct {
	Multihost []Request `json:"multihost"`
}
type MultiInfo struct {
	ResultArray []CBMCISMetric `json:"resultarray"`
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

var Content CBMCISMetric

func NewErrMSG(message string) map[string]string {
	errMsg := map[string]string{
		"message": message,
	}
	return errMsg
}
