package telemetry

import (
	"fmt"
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

const DefaultJaegerCollectorEndpoint = "http://localhost:14268/api/traces"

func Init(serviceName string, config Config) (FlushFunc, error) {
	var (
		tp    trace.Provider
		flush FlushFunc
		err   error
	)

	SetServiceName(serviceName)

	if !config.Enabled {
		tp = trace.NoopProvider{}
	} else {
		switch config.Exporter.Type {
		case "jaeger":
			jaegerConfig := config.Exporter.Jaeger
			tp, flush, err = jaeger.NewExportPipeline(
				jaeger.WithCollectorEndpoint(jaegerConfig.CollectorEndpoint),
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

		default:
			return nil, fmt.Errorf("unsupported exporter type: %s", config.Exporter.Type)
		}
	}

	global.SetTraceProvider(tp)
	return flush, nil
}
