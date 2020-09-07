package cbagent

import (
	"context"
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
}

// API 서버 생성
func NewAgentAPIServer(port int) AgentAPIServer {
	return AgentAPIServer{
		listenPort: port,
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

	//통합 매트릭 수집
	g.GET("/metric", server.getAllMetric)

	//선택 매트릭 수집
	g.GET("/metric/:type", server.getMetric)

	//TTL 수집
	g.GET("/cmd", server.performCommand)

	//동작
	if err := e.Start(":8080"); err != nil {
		logrus.Fatal(err)
	}
	return nil
}

//TODO: MCIS Monitoring 협의 후 결정
func (server *AgentAPIServer) performCommand(c echo.Context) error {
	return nil
}
