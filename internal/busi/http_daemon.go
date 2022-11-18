package busi

import (
	v1 "data-extraction-notify/internal/busi/api/v1"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setWalkerConfig(Lotus0, Mq string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(v1.LOTUS0, Lotus0)
		c.Set(v1.MQ, Mq)

		c.Next()
	}
}

func (s *HttpServer) registerV1(r *gin.Engine) {
	apiv1 := r.Group("/api/v1")
	{
		apiv1.GET("/ping", v1.Ping)

		apiv1.POST("/topic", v1.TopicSignIn)            //-CallByTaskModel
		apiv1.DELETE("/topic", v1.TopicDelete)          //-CallByTaskModel
		apiv1.POST("/task_state", v1.ReportTipsetState) //-CallByTaskModel

		apiv1.POST("/walk", setWalkerConfig(s.lotus0, s.mq), v1.WalkTipsets) //-CallByManual

		apiv1.POST("/gapfill", setWalkerConfig(s.lotus0, s.mq), v1.GapFill)     //-CallByScheduler
		apiv1.POST("/retry", setWalkerConfig(s.lotus0, s.mq), v1.ReplayTipsets) //-CallByScheduler
	}
}

func (s *HttpServer) RegisterRoutes(r *gin.Engine) {
	// r.Use(utils.Cors())
	r.Use(cors.Default())
	r.GET("/data-extraction-notify/swagger/*any", swagHandler)

	s.registerV1(r)
}

func (s *HttpServer) Start() {
	// // if Flags.Mode == "prod" {
	gin.SetMode(gin.ReleaseMode)
	// // }

	// r := gin.Default()
	r := gin.New()
	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	r.Use(gin.Recovery())
	s.RegisterRoutes(r)
	r.Run(s.addr)
}

type HttpServer struct {
	addr   string
	lotus0 string
	mq     string
}

func NewHttpServer(addr string, lotus string, mq string) *HttpServer {
	return &HttpServer{addr, lotus, mq}
}

func HttpServerStart(addr string, lotus string, mq string) {
	NewHttpServer(addr, lotus, mq).Start()
}
