package main

import (
	"kubescape-config-service/utils"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//////////////////////////////////////////middleware handlers//////////////////////////////////////////

// authenticate middleware for request authentication
func authenticate(c *gin.Context) {
	customerGuid, err := c.Cookie(utils.CUSTOMER_GUID)
	if err != nil || customerGuid == "" {
		customerGuid := c.Query(utils.CUSTOMER_GUID)
		if customerGuid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	}
	c.Set(utils.CUSTOMER_GUID, customerGuid)
	c.Next()
}

// traceAttributesNHeader middleware adds tracing header in response and request attributes in span
func traceAttributesNHeader(c *gin.Context) {
	otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.HeaderCarrier(c.Writer.Header()))
	if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
		trace.SpanFromContext(c.Request.Context()).SetAttributes(attribute.String(utils.CUSTOMER_GUID, c.GetString(utils.CUSTOMER_GUID)))
	}

	c.Next()
}

// requestLogger middleware adds a logger with request attributes to the context
func requestLoggerWithFields(c *gin.Context) {
	fields := []zapcore.Field{
		zap.String("method", c.Request.Method),
		zap.String("query", c.Request.URL.Path),
		zap.String("path", c.Request.URL.RawQuery),
	}
	fields = append(fields, telemetryLogFields(c)...)
	c.Set("zapLogger", zapLogger.WithOptions(zap.Fields(fields...)))
	c.Next()
}

// requestSummary middleware logs request summary after request is served
func requestSummary() func(c *gin.Context) {
	return ginzap.GinzapWithConfig(zapInfoLevelLogger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context:    telemetryLogFields,
	})
}

/////////////////////////////////////helper functions/////////////////////////////////////

// telemetryLogFields returns telemetry and customer id fields for  logging
func telemetryLogFields(c *gin.Context) []zapcore.Field {
	fields := []zapcore.Field{}
	// log request ID
	if customerGUID := c.GetString(utils.CUSTOMER_GUID); customerGUID != "" {
		fields = append(fields, zap.String(utils.CUSTOMER_GUID, customerGUID))
	}
	// log trace and span ID
	if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
		fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
		fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
	}
	return fields
}
