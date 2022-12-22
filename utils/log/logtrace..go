package log

import (
	"config-service/utils/consts"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LogNTrace(msg string, c context.Context, fields ...zapcore.Field) {
	Log(msg, c, fields...)
	AddEvent(msg, c)
}

// LogNTraceEnterExit logs the entry and exit of a function
// - usage: defer LogNTraceEnterExit("function name", c)()
// output: function name  // entry
// output: function name completed // exit
func LogNTraceEnterExit(msg string, c context.Context, fields ...zapcore.Field) func() {
	LogNTrace(msg, c, fields...)
	return func() {
		LogNTrace(msg+" completed", c, fields...)
	}
}

func LogNTraceError(msg string, err error, c context.Context, fields ...zapcore.Field) {
	LogError(msg, err, c, fields...)
	AddEvent(msg, c)
	TraceErrorSetStatus(msg, err, c)
	if err != nil {
		if g, ok := c.(*gin.Context); ok {
			g.Error(err)
		}
	}
}

func Log(msg string, c context.Context, fields ...zapcore.Field) {
	doLog(msg, nil, c, fields...)
}

func LogError(msg string, err error, c context.Context, fields ...zapcore.Field) {
	doLog(msg, err, c, fields...)
}

func AddEvent(msg string, c context.Context) {
	if span := GetTraceSpan(c); span != nil {
		span.AddEvent(msg)
	}
}

func TraceErrorSetStatus(msg string, err error, c context.Context) {
	if span := GetTraceSpan(c); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("%s Error: %v", msg, err))
	}
}

func GetLogger(c context.Context) *zap.Logger {
	if z := c.Value(consts.ReqLogger); z != nil {
		return z.(*zap.Logger)
	}
	return zap.L()
}

func GetTraceSpan(c context.Context) trace.Span {
	if g, ok := c.(*gin.Context); ok {
		if trace.SpanFromContext(g.Request.Context()).SpanContext().IsValid() {
			return trace.SpanFromContext(g.Request.Context())
		}
	}
	return nil
}

func doLog(msg string, err error, c context.Context, fields ...zapcore.Field) {
	logger := GetLogger(c)
	if err != nil {
		logger.Error(msg, append(fields, zap.Error(err))...)
	} else {
		logger.Info(msg, fields...)
	}
}
