package listener

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/influxdata/telegraf"
	runagent "github.com/influxdata/telegraf/cloudbarista/cbagent"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	common "github.com/influxdata/telegraf/cloudbarista/utility"
	"github.com/labstack/echo"
)

// doHealthCheck 에이전트 헬스체크
func (server *AgentPullLister) doHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusNoContent, nil)
}

// getMetric 온디맨드 모니터링 선택 메트릭 수집
func (server *AgentPullLister) getMetric(c echo.Context) error {
	//쿼리에서 수집 메트릭 정보 파싱
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	//전체 매트릭 수집
	value, err := server.gatherMetric()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Telegraf 매트릭 DragonFly 매트릭으로 변환
	convertedMetric := usage.ToCBMetric(value)

	switch metrictype {
	// 통합 메트릭 응답
	case "all":
		log.Println("Start Closing Cloud-Barista Agent")
		log.Println("===============================================================================================================================")
		return c.JSON(http.StatusOK, convertedMetric)
	default:
		// 선택 메트릭 추출 후 응답
		result := usage.ExtractMetric(metrictype, convertedMetric)
		log.Println("Start Closing Cloud-Barista Agent")
		log.Println("===============================================================================================================================")
		if len(result) == 0 {
			return c.JSON(http.StatusBadRequest, "Not Found Metric")
		}
		return c.JSON(http.StatusOK, result)
	}

}

// gatherMetric 메트릭 수집을 위한 에이전트 동작
func (server *AgentPullLister) gatherMetric() (map[string]telegraf.Metric, error) {
	result, err := runagent.RunAgent(*server.Ctx, usage.InputFilters, usage.OutputFilters)
	if err != nil {
		err = errors.New("Running CBAgent Error")
	}
	return result, err
}

func (server *AgentPullLister) mcisMetric(c echo.Context) error {
	mcis.CleanMCISMetric()
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	isServed := false
	for _, value := range server.McisMetric {
		if metrictype == value {
			isServed = true
		}
	}
	if !isServed {
		return c.JSON(http.StatusBadRequest, "Not Found Metric")
	}
	reflect.ValueOf(server.MCISAgent[mcis.MCIS]).MethodByName(metrictype).Call([]reflect.Value{reflect.ValueOf(c)})
	return nil
}

func (server *AgentPullLister) handlePushMonitoring(c echo.Context) error {
	switch c.Request().Method {
	case common.METHOD_CREATE:
		if server.IsPushModuleOn {
			return c.JSON(http.StatusInternalServerError, "Push Monitoring Already Activated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Starting Cloud-Barista Push Monitoring Telegraf Agent")
		server.IsPushModuleOn = true
		server.pushControllerChan <- true
		return c.JSON(http.StatusOK, fmt.Sprintf("Push Monitoring Started && Pull Monitoring Stopped"))
	case common.METHOD_DELETE:
		if !server.IsPushModuleOn {
			return c.JSON(http.StatusInternalServerError, "Push Monitoring Already DeActivated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Stopping Cloud-Barista Push Monitoring Telegraf Agent")
		server.IsPushModuleOn = false
		server.pushControllerChan <- false
		return c.JSON(http.StatusOK, fmt.Sprintf("Push Monitoring Stopped && Pull Monitoring Started"))
	default:
		return c.JSON(http.StatusBadRequest, "Unsupported Request")
	}
}
