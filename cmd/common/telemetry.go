package common

import (
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/spf13/cobra"
)

func GetTelemetryConfig(cmd *cobra.Command) (telemetry.Config, error) {
	tracingEnabled, err := cmd.PersistentFlags().GetBool("tracing.enabled")
	if err != nil {
		return telemetry.Config{}, err
	}

	jaegerCollectorURL, err := cmd.PersistentFlags().GetString("tracing.jaeger-collector-url")
	if err != nil {
		return telemetry.Config{}, err
	}

	cfg := telemetry.Config{
		Enabled:  tracingEnabled,
		Exporter: telemetry.ExporterConfig{
			Type:   "jaeger",
			Jaeger: telemetry.JaegerExporterConfig{
				CollectorEndpoint: jaegerCollectorURL,
			},
		},
	}

	return cfg, nil
}
