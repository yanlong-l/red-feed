package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *Builder {
	return &Builder{
		loggerFunc: fn,
	}
}

func (b *Builder) AllowReqBody(allow bool) *Builder {
	b.allowReqBody = allow
	return b
}

func (b *Builder) AllowRespBody(allow bool) *Builder {
	b.allowRespBody = allow
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		al := &AccessLog{
			Path:   ctx.Request.URL.Path,
			Method: ctx.Request.Method,
		}
		var reqBody []byte
		if b.allowReqBody && ctx.Request.Body != nil {
			reqBody, _ = ctx.GetRawData()
			al.ReqBody = string(reqBody)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		if b.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             al,
			}
		}

		defer func() {
			al.Duration = time.Since(startTime).String()
			b.loggerFunc(ctx, al)
		}()

		// 执行业务逻辑
		ctx.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(code int) {
	w.al.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

type AccessLog struct {
	Path       string
	Method     string
	Duration   string
	StatusCode int
	ReqBody    string
	RespBody   string
}
