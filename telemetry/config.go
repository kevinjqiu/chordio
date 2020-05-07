package telemetry

type JaegerExporterConfig struct {
	CollectorEndpoint string `mapstructure:"collectorEndpoint"`
}

type ExporterConfig struct {
	Type   string               `mapstructure:"type"`
	Jaeger JaegerExporterConfig `mapstructure:"jaeger"`
}

type Config struct {
	Enabled  bool `mapstructure:"enabled"`
	Exporter ExporterConfig
}
