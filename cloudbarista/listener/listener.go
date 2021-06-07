package listener

import (
	"context"
	cbutility "github.com/influxdata/telegraf/cloudbarista/utility"
	"net/http"
	"os"

	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

type AgentPullListener struct {
	Echo           *echo.Echo
	listenPort     int
	MCISAgent      map[string]interface{}
	pushModuleChan chan bool
	IsPushOn       bool
	signals        chan os.Signal
	Ctx            *context.Context
	Cancel         context.CancelFunc
	McisMetric     map[string]string
	pushCheckChan  chan bool
}

// Listener 서버 생성
func NewAgentPullListener(port int, pushModuleChan chan bool, signals chan os.Signal, pushCheckChan chan bool) AgentPullListener {
	var mcisAgent = map[string]interface{}{
		mcis.MCIS: &mcis.MCISAgent{},
	}
	var emptyMcisMetric = map[string]string{}
	e := echo.New()
	ctx, cancel := context.WithCancel(context.Background())
	var newAgentPullListener = AgentPullListener{
		Echo:           e,
		listenPort:     port,
		MCISAgent:      mcisAgent,
		pushModuleChan: pushModuleChan,
		IsPushOn:       cbutility.OFF,
		signals:        signals,
		Ctx:            &ctx,
		Cancel:         cancel,
		McisMetric:     emptyMcisMetric,
		pushCheckChan:  pushCheckChan,
	}
	mcis.InitializeMetricList(newAgentPullListener.McisMetric)
	return newAgentPullListener
}

// Listener 동작
func (listener *AgentPullListener) Start() error {
	listener.Echo.Use(middleware.Logger())
	listener.Echo.Use(middleware.Recover())
	listener.Echo.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cloud-Barista Telegraf Agent API Server")
	})

	// API 그룹화
	g := listener.Echo.Group("/cb-dragonfly")

	// 헬스체크
	g.GET("/healthcheck", listener.doHealthCheck)

	// 온디맨드 모니터링 매트릭 수집
	g.GET("/metric/:metric_name", listener.getMetric)

	// TTL 수집
	g.GET("/mcis/metric/:metric_name", listener.mcisMetric)

	// Push Monitoring On
	g.POST("/agent/monitoring/push", listener.handlePushMonitoring)
	// Push Monitoring Off
	g.DELETE("/agent/monitoring/push", listener.handlePushMonitoring)

	//동작
	go func() {
		if err := listener.Echo.Start(":8888"); err != nil {
			logrus.Fatal(err)
		}
	}()
	for {
		select {
		case _ = <-listener.signals:
			_ = listener.Shutdown()
		}
	}
}

func (listener *AgentPullListener) Shutdown() error {
	defer listener.Cancel()
	if err := listener.Echo.Shutdown(*listener.Ctx); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
