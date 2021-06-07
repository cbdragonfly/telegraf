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
	cbUtility "github.com/influxdata/telegraf/cloudbarista/utility"
	common "github.com/influxdata/telegraf/cloudbarista/utility"
	"github.com/labstack/echo"
)

// doHealthCheck 에이전트 헬스체크
func (listener *AgentPullListener) doHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusNoContent, nil)
}

// getMetric 온디맨드 모니터링 선택 메트릭 수집
func (listener *AgentPullListener) getMetric(c echo.Context) error {
	//쿼리에서 수집 메트릭 정보 파싱
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	//전체 매트릭 수집
	value, err := listener.gatherMetric()
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
func (listener *AgentPullListener) gatherMetric() (map[string]telegraf.Metric, error) {
	result, err := runagent.RunAgent(*listener.Ctx, usage.InputFilters, usage.OutputFilters)
	if err != nil {
		err = errors.New("Running CBAgent Error")
	}
	return result, err
}

func (listener *AgentPullListener) mcisMetric(c echo.Context) error {
	mcis.CleanMCISMetric()
	metrictype := c.Param("metric_name")
	if metrictype == "" {
		err := errors.New("Failed to get metrictype from query")
		return c.JSON(http.StatusInternalServerError, err)
	}
	isServed := false
	for _, value := range listener.McisMetric {
		if metrictype == value {
			isServed = true
		}
	}
	if !isServed {
		return c.JSON(http.StatusBadRequest, "Not Found Metric")
	}
	reflect.ValueOf(listener.MCISAgent[mcis.MCIS]).MethodByName(metrictype).Call([]reflect.Value{reflect.ValueOf(c)})
	return nil
}

func (listener *AgentPullListener) handlePushMonitoring(c echo.Context) error {
	switch c.Request().Method {
	case common.METHOD_CREATE:
		if listener.IsPushOn {
			return c.JSON(http.StatusForbidden, "Push Monitoring Already Activated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Starting Cloud-Barista Push Monitoring Telegraf Agent")
		listener.pushModuleChan <- cbUtility.ON
		break
	case common.METHOD_DELETE:
		if !listener.IsPushOn {
			return c.JSON(http.StatusForbidden, "Push Monitoring Already DeActivated")
		}
		log.Println("===============================================================================================================================")
		log.Printf("I! Stopping Cloud-Barista Push Monitoring Telegraf Agent")
		listener.pushModuleChan <- cbUtility.OFF
		break
	default:
		return c.JSON(http.StatusBadRequest, "Unsupported Request")
	}
	return listener.handlePushCheckResponse(c)
}

func (listener *AgentPullListener) handlePushCheckResponse(c echo.Context) error {
	for {
		select {
		case pushChecker := <-listener.pushCheckChan:
			listener.IsPushOn = pushChecker
			return c.JSON(http.StatusOK, fmt.Sprintf("Push Monitoring %s API completed", c.Request().Method))
		}
	}
}
