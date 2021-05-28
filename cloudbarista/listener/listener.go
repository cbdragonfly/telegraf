package listener

import (
	"context"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type AgentPullLister struct {
	Echo               *echo.Echo
	listenPort         int
	MCISAgent          map[string]interface{}
	pushControllerChan chan bool
	IsPushModuleOn     bool
	signals            chan os.Signal
	Ctx                *context.Context
	Cancel             context.CancelFunc
	McisMetric         map[string]string
}

// Listener 서버 생성
func NewAgentPullListener(port int, pushControllerChan chan bool, signals chan os.Signal) AgentPullLister {
	var mcisAgent = map[string]interface{}{
		mcis.MCIS: &mcis.MCISAgent{},
	}
	var emptyMcisMetric = map[string]string{}
	e := echo.New()
	ctx, cancel := context.WithCancel(context.Background())
	var newAgentPullListener = AgentPullLister{
		Echo:               e,
		listenPort:         port,
		MCISAgent:          mcisAgent,
		pushControllerChan: pushControllerChan,
		IsPushModuleOn:     false,
		signals:            signals,
		Ctx:                &ctx,
		Cancel:             cancel,
		McisMetric:         emptyMcisMetric,
	}
	mcis.InitializeMetricList(newAgentPullListener.McisMetric)
	return newAgentPullListener
}

// Listener 동작
func (listener *AgentPullLister) Start() error {
	listener.Echo.Use(middleware.Logger())
	listener.Echo.Use(middleware.Recover())
	listener.Echo.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cloud-Barista Telegraf Agent API Server")
	})
	// API 그룹화
	g := listener.Echo.Group("/cb-dragonfly")

	// 온디맨드 모니터링 매트릭 수집
	g.GET("/metric/:metric_name", listener.getMetric)

	//TTL 수집
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

func (listener *AgentPullLister) Shutdown() error {
	defer listener.Cancel()
	if err := listener.Echo.Shutdown(*listener.Ctx); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
