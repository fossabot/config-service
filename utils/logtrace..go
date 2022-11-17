package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LogNTrace(msg string, c *gin.Context, fields ...zapcore.Field) {
	Log(msg, c, fields...)
	AddEvent(msg, c)
}

func LogNTraceError(msg string, err error, c *gin.Context, fields ...zapcore.Field) {
	LogError(msg, err, c, fields...)
	AddEvent(msg, c)
	TraceErrorSetStatus(msg, err, c)
}

func Log(msg string, c *gin.Context, fields ...zapcore.Field) {
	doLog(msg, nil, c, fields...)
}

func LogError(msg string, err error, c *gin.Context, fields ...zapcore.Field) {
	doLog(msg, err, c, fields...)
}

func AddEvent(msg string, c *gin.Context) {
	if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
		trace.SpanFromContext(c.Request.Context()).AddEvent(msg)
	}
}

func TraceErrorSetStatus(msg string, err error, c *gin.Context) {
	if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
		trace.SpanFromContext(c.Request.Context()).SetStatus(codes.Error, fmt.Sprintf("%s Error: %v", msg, err))
	}
}

func GetLogger(c *gin.Context) *zap.Logger {
	if z, ok := c.Get("zapLogger"); ok {
		return z.(*zap.Logger)
	}
	return zap.L()
}

func GetTraceSpan(c *gin.Context) trace.Span {
	if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
		return trace.SpanFromContext(c.Request.Context())
	}
	return nil
}

func doLog(msg string, err error, c *gin.Context, fields ...zapcore.Field) {
	logger := GetLogger(c)
	if err != nil {
		logger.Error(msg, append(fields, zap.Error(err))...)
	} else {
		logger.Info(msg, fields...)
	}
}
