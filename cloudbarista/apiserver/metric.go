package cbagent

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	runagent "github.com/influxdata/telegraf/cloudbarista/cbagent"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/influxdata/telegraf/cloudbarista/push"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"reflect"
)

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

func (server *AgentAPIServer) mcisMetric(c echo.Context) error {
	mcis.CleanMCISMetric()
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	reflect.ValueOf(server.MCISAgent[mcis.MCIS]).MethodByName(metrictype).Call([]reflect.Value{reflect.ValueOf(c)})
	return nil
}

func (server *AgentAPIServer) handlePushMonitoring(c echo.Context) error {
	switch c.Request().Method {
	case push.METHOD_CREATE:
		if server.isPushModuleOn {
			return c.JSON(http.StatusInternalServerError, "Push Monitoring Already Activated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Starting Cloud-Barista Push Monitoring Telegraf Agent")
		server.isPushModuleOn = true
		server.pushControllerChan <- true
		return c.JSON(http.StatusOK, fmt.Sprintf("Push Monitoring Started && Pull Monitoring Stopped"))
	default:
		if !server.isPushModuleOn {
			return c.JSON(http.StatusInternalServerError, "Push Monitoring Already DeActivated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Stopping Cloud-Barista Push Monitoring Telegraf Agent")
		server.isPushModuleOn = false
		server.pushControllerChan <- false
		return c.JSON(http.StatusOK, fmt.Sprintf("Push Monitoring Stopped && Pull Monitoring Started"))
	}
}
