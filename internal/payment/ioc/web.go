package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/lazywoo/mercury/internal/payment/web"
	"github.com/lazywoo/mercury/pkg/ginx"
)

func InitWebServer(hdl *web.WechatHandler) *ginx.Server {
	engine := gin.Default()
	hdl.RegisterRoutes(engine)
	addr := viper.GetString("http.addr")
	ginx.InitCounterVec(prometheus.CounterOpts{
		Namespace: "mercury",
		Subsystem: "payment",
		Name:      "http",
	})
	return &ginx.Server{
		Addr:   addr,
		Engine: engine,
	}
}
