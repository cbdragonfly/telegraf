package cbagent

import (
	"errors"
	"github.com/influxdata/telegraf"
	"log"
	"net/http"
	"sync"

	runagent "github.com/influxdata/telegraf/cloudbarista/cbagent"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/labstack/echo"
)

var wg sync.WaitGroup

//온디맨드 모니터링 선택 메트릭 수집
func (server *AgentAPIServer) getMetric(c echo.Context) error {
	//쿼리에서 수집 메트릭 정보 파싱
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	//전체 매트릭 수집
	value, err := gatherMetric()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Telegraf 매트릭 DragonFly 매트릭으로 변환
	convertedMetric := usage.ToCBMetric(value)

	wg.Wait()

	switch metrictype {
	// 통합 메트릭 응답
	case "all":
		log.Println("Closing Cloud-Barista Agent")
		log.Println("===============================================================================================================================")
		return c.JSON(http.StatusOK, convertedMetric)
	default:
		// 선택 메트릭 추출 후 응답
		result := usage.ExtractMetric(metrictype, convertedMetric)
		log.Println("Closing Cloud-Barista Agent")
		log.Println("===============================================================================================================================")
		return c.JSON(http.StatusOK, result)
	}

}

// 메트릭 수집을 위한 에이전트 동작
func gatherMetric() (map[string]telegraf.Metric, error) {
	result, err := runagent.RunAgent(usage.Ctx, usage.InputFilters, usage.OutputFilters)
	if err != nil {
		err = errors.New("Running CBAgent Error")
	}
	return result, err
}
