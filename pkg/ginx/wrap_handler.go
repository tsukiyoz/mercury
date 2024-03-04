package ginx

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tsukaychan/webook/pkg/logger"
)

var log logger.Logger = logger.NewNopLogger()

var counterVec *prometheus.CounterVec

func InitCounterVec(opt prometheus.CounterOpts) {
	counterVec = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(counterVec)
}

func SetLogger(l logger.Logger) {
	log = l
}

func WrapReq[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}
		res, err := fn(ctx, req)
		if err != nil {
			log.Error("processing business logic error",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err),
			)
		}
		counterVec.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[Claims jwt.Claims](fn func(ctx *gin.Context, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := val.(*Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, *claims)
		if err != nil {
			log.Error("processing business logic error",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err),
			)
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaimsAndReq[Req any, Claims jwt.Claims](fn func(ctx *gin.Context, req Req, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := val.(*Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, req, *claims)
		if err != nil {
			log.Error("processing business logic error",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err),
			)
		}
		ctx.JSON(http.StatusOK, res)
	}
}
