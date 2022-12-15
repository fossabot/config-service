//go:build go1.8
// +build go1.8

package main

import (
	"config-service/db/mongo"
	"config-service/utils"
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	jaegerexporter "go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLogger *zap.Logger
var zapInfoLevelLogger *zap.Logger

func initialize() (shutdown func()) {
	conf := utils.GetConfig()
	//init logger
	initLogger(conf.LoggerConfig)
	//init tracer
	tracer := initTracer(conf.Telemetry)
	//connect db
	mongo.MustConnect(conf.Mongo)

	//shutdown function
	shutdown = func() {
		mongo.Disconnect()
		if err := tracer.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
		zapLogger.Sync()
	}
	return shutdown
}

func initLogger(config utils.LoggerConfig) {
	var err error
	lvl := zap.NewAtomicLevel()
	if config.Level == "" {
		config.Level = "warn"
	}
	if err := lvl.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}
	// empty string means the output is stdout
	zapConf := newZapConf(lvl, config)
	zapLogger, err = zapConf.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(zapLogger)
	zap.RedirectStdLog(zapLogger)
	if lvl.Level() == zapcore.InfoLevel {
		zapInfoLevelLogger = zapLogger
	} else {
		zapInfoLevelLogger, err = newZapConf(zap.NewAtomicLevelAt(zapcore.InfoLevel), config).Build()
		if err != nil {
			panic(err)
		}
	}
}

func newZapConf(lvl zap.AtomicLevel, config utils.LoggerConfig) zap.Config {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	zapConf := zap.Config{DisableCaller: true, DisableStacktrace: true, Level: lvl,
		Encoding: "json", EncoderConfig: ec,
		OutputPaths: []string{"stdout"}, ErrorOutputPaths: []string{"stderr"}}
	if config.LogFileName != "" {
		zapConf.OutputPaths = []string{config.LogFileName}
		zapConf.ErrorOutputPaths = append(zapConf.ErrorOutputPaths, config.LogFileName)
	}
	return zapConf
}

// initTracer used to initialize tracer
func initTracer(config utils.TelemetryConfig) *sdktrace.TracerProvider {
	serviceName := "config-service"
	hostName, _ := os.Hostname()
	var err error
	var exporter sdktrace.SpanExporter
	if config.JaegerAgentHost != "" && config.JaegerAgentPort != "" {
		if config.JaegerAgentHost == "stdout" {
			exporter, err = stdout.New(stdout.WithPrettyPrint())
			if err != nil {
				log.Fatal(err)
			}

		} else {
			host := jaegerexporter.WithAgentHost(config.JaegerAgentHost)
			port := jaegerexporter.WithAgentPort(config.JaegerAgentPort)
			// jaegerexporter.With
			exporter, err = jaegerexporter.New(jaegerexporter.WithAgentEndpoint(host, port))
			if err != nil {
				zapLogger.Fatal("failed to initialize jaeger exporter", zap.Error(err))
			}
		}
	}
	var tp *sdktrace.TracerProvider
	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("pod", hostName),
			semconv.K8SPodNameKey.String(hostName))),
	}
	if exporter != nil {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}
	tp = sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
