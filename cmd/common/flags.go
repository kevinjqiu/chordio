package common

import (
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/spf13/cobra"
)

type (
	CommonFlags struct {
		Loglevel string
		Tracing  TracingConfig
	}

	TracingConfig struct {
		Enabled            bool
		JaegerCollectorURL string
	}
)

func AddCommonPflags(cmd *cobra.Command, flags *CommonFlags) {
	cmd.PersistentFlags().StringVarP(&flags.Loglevel, "loglevel", "l", "info", "log level")
	cmd.PersistentFlags().BoolVarP(&flags.Tracing.Enabled, "tracing.enabled", "t", true, "enable opentracing")
	cmd.PersistentFlags().StringVarP(&flags.Tracing.JaegerCollectorURL, "tracing.jaeger-collector-url", "r", telemetry.DefaultJaegerCollectorEndpoint, "jaeger collector URL")
}

