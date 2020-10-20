package cbagent

import (
	"context"
	"github.com/influxdata/telegraf/cloudbarista/mcis"
	"net/http"

	cbagent "github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/cloudbarista/usage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

type AgentAPIServer struct {
	listenPort int
	a          *cbagent.Agent
	MCISAgent  map[string]interface{}
}

// API 서버 생성
func NewAgentAPIServer(port int) AgentAPIServer {
	var mcisAgent = map[string]interface{}{
		mcis.MCIS: &mcis.MCISAgent{},
	}
	return AgentAPIServer{
		listenPort: port,
		MCISAgent:  mcisAgent,
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

	//동작
	if err := e.Start(":8080"); err != nil {
		logrus.Fatal(err)
	}
	return nil
}
