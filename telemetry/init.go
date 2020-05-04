package telemetry

import (
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type FlushFunc func()

func Init(config Config) (FlushFunc, error) {
	//exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	//if err != nil {
	//	return err
	//}
	//tp, err := sdktrace.NewProvider(
	//	sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	//	sdktrace.WithSyncer(exporter),
	//)
	//if err != nil {
	//	return err
	//}

	tp, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "chordio",
			Tags: []core.KeyValue{
				key.String("exporter", "jaeger"),
			},
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)

	if err != nil {
		return nil, err
	}

	global.SetTraceProvider(tp)
	return flush, nil
}
