package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	LogFn         func(ctx context.Context, l LogContent) // 传入一个方法控制如何打印日志
	allowReqBody  bool                                    // 定义一个控制参数，防止黑客传入一个很大的 Body，打崩日志系统
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l LogContent)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		LogFn: logFn,
	}

}

// 链式调用
func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

// 链式调用
func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 拿到请求

		path := ctx.Request.URL.Path

		if len(path) > 1024 { // 防止黑客传入一个很长的路径，进行截断
			path = path[:1024]
		}

		method := ctx.Request.Method

		lc := LogContent{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody { // 允许打印请求体

			// RequestBody 是一个 Stream 对象，读出来，需要再放回去
			body, _ := ctx.GetRawData() // 拿到请求体
			// 可以考虑忽略错误
			if len(body) > 2048 { // 控制 body 大小
				lc.ReqBody = string(body[:2048])
			} else {
				lc.ReqBody = string(body)
			}
			// 将 body 放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		start := time.Now()

		if l.allowRespBody {
			// ctx 中没有 Response，那么怎么拿到 RepsBody
			// 原始的 ctx.Writer 是拿不到响应数据的，那么我们可以通过装饰器 重新实现一下 ResponseWriter，拿到响应数据
			ctx.Writer = &responseWriter{
				lc:             &lc,
				ResponseWriter: ctx.Writer,
			}
		}

		// 拿到响应
		defer func() {
			lc.Duration = time.Since(start) // 执行时间
			l.LogFn(ctx, lc)

		}()

		// 执行下一个 Middleware...直到业务逻辑
		ctx.Next()

		// 在这里通过 defer 就拿到了响应
	}

}

type LogContent struct { // 请求和响应的日志内容
	Path       string        `json:"path"`      // 请求路径
	Method     string        `json:"method"`    // 请求方法
	ReqBody    string        `json:"req_body"`  // 请求体
	RespBody   string        `json:"resp_body"` // 响应体
	Duration   time.Duration `json:"duration"`  // 执行时间
	StatusCode int
}

type responseWriter struct {
	gin.ResponseWriter
	lc *LogContent
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.lc.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.lc.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
