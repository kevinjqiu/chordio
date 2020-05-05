package telemetry

import (
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	telemetryServiceName = ""
)

func SetServiceName(name string) {
	telemetryServiceName = name
}

func GetServiceName() string {
	return telemetryServiceName
}

type FlushFunc func()

const jaegerCollectorEndpoint = "http://localhost:14268/api/traces"

func Init(serviceName string, config Config) (FlushFunc, error) {
	var (
		tp    trace.Provider
		flush FlushFunc
		err   error
	)

	SetServiceName(serviceName)

	if config.Enabled {
		tp = trace.NoopProvider{}
	} else {
		tp, flush, err = jaeger.NewExportPipeline(
			jaeger.WithCollectorEndpoint(jaegerCollectorEndpoint),
			jaeger.WithProcess(jaeger.Process{
				ServiceName: serviceName,
				Tags: []core.KeyValue{
					key.String("exporter", "jaeger"),
				},
			}),
			jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		)

		if err != nil {
			return nil, err
		}
	}

	global.SetTraceProvider(tp)
	return flush, nil
}
