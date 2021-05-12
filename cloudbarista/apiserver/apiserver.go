package cbagent

import (
	"context"
	cbagent "github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type AgentAPIServer struct {
	listenPort         int
	a                  *cbagent.Agent
	MCISAgent          map[string]interface{}
	pushControllerChan chan bool
	isPushModuleOn     bool
}

// API 서버 생성
func NewAgentAPIServer(port int, pushControllerChan chan bool) AgentAPIServer {
	var mcisAgent = map[string]interface{}{
		mcis.MCIS: &mcis.MCISAgent{},
	}
	return AgentAPIServer{
		listenPort:         port,
		MCISAgent:          mcisAgent,
		pushControllerChan: pushControllerChan,
		isPushModuleOn:     false,
	}
}

// API 서버 동작

func (server *AgentAPIServer) RunAPIServer() error {

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
	g.GET("/metric/:metric_name", server.getMetric)

	//TTL 수집
	g.GET("/mcis/metric/:metric_name", server.mcisMetric)

	// Push Monitoring On
	g.POST("/agent/monitoring/push", server.handlePushMonitoring)
	// Push Monitoring Off
	g.DELETE("/agent/monitoring/push", server.handlePushMonitoring)
	//동작
	go func() {
		if err := e.Start(":8888"); err != nil {
			logrus.Fatal(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGTERM, syscall.SIGINT)
	<-signals

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
