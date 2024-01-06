package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type MiddlewareBuilder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, aL *AccessLog)
}

func NewMiddlewareBuilder(fn func(ctx context.Context, aL *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc: fn,
	}
}

type AccessLog struct {
	Method   string
	Url      string
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}

func (b *MiddlewareBuilder) AllowReqBody() *MiddlewareBuilder {
	b.allowReqBody = true
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		accessLog := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if b.allowReqBody && ctx.Request.Body != nil {
			reqBody, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

			if len(reqBody) > 1024 {
				reqBody = reqBody[:1024]
			}

			accessLog.ReqBody = string(reqBody)
		}

		if b.allowRespBody {
			ctx.Writer = responseWriter{
				accessLog:      accessLog,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			accessLog.Duration = time.Since(start).String()
			b.loggerFunc(ctx, accessLog)
		}()

		// execute business logic
		ctx.Next()
	}
}

type responseWriter struct {
	accessLog *AccessLog
	gin.ResponseWriter
}

func (w responseWriter) Write(data []byte) (int, error) {
	w.accessLog.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.accessLog.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.accessLog.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
