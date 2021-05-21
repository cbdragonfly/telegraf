package listener

import (
	"context"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"syscall"
)

type AgentPullLister struct {
	listenPort         int
	MCISAgent          map[string]interface{}
	pushControllerChan chan bool
	IsPushModuleOn     bool
	signals            chan os.Signal
	ctx                context.Context
	cancel             context.CancelFunc
}

// Listener 서버 생성
func NewAgentPullListener(port int, pushControllerChan chan bool, signals chan os.Signal, ctx context.Context, cancel context.CancelFunc) AgentPullLister {
	var mcisAgent = map[string]interface{}{
		mcis.MCIS: &mcis.MCISAgent{},
	}
	return AgentPullLister{
		listenPort:         port,
		MCISAgent:          mcisAgent,
		pushControllerChan: pushControllerChan,
		IsPushModuleOn:     false,
		signals:            signals,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Listener 동작
func (listener *AgentPullLister) Start() error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	usage.Ctx = context.Background()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cloud-Barista Telegraf Agent API Server")
	})
	// API 그룹화
	g := e.Group("/cb-dragonfly")

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
		if err := e.Start(":8888"); err != nil {
			logrus.Fatal(err)
		}
	}()

	defer listener.cancel()
	for {
		select {
		case signal := <-listener.signals:
			if signal != syscall.SIGHUP {
				if err := e.Shutdown(listener.ctx); err != nil {
					logrus.Fatal(err)
				}
			}
		}
	}
}
