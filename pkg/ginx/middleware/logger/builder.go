package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

type MiddlewareBuilder struct {
	allowReqBody  *atomic.Bool
	allowRespBody *atomic.Bool
	loggerFunc    func(ctx context.Context, aL *AccessLog)
}

func NewMiddlewareBuilder(fn func(ctx context.Context, aL *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc:    fn,
		allowReqBody:  atomic.NewBool(false),
		allowRespBody: atomic.NewBool(false),
	}
}

func (b *MiddlewareBuilder) AllowReqBody(allow bool) *MiddlewareBuilder {
	b.allowReqBody.Store(allow)
	return b
}

func (b *MiddlewareBuilder) AllowRespBody(allow bool) *MiddlewareBuilder {
	b.allowRespBody.Store(allow)
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		accessLog := NewAccessLog(ctx.Request.Method, limitString(ctx.Request.URL.String()))

		if b.allowReqBody.Load() && ctx.Request.Body != nil {
			reqBody, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

			accessLog.ReqBody = string(limitByte(reqBody))
		}

		if b.allowRespBody.Load() {
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
	data = limitByte(data)
	w.accessLog.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	data = limitString(data)
	w.accessLog.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.accessLog.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
