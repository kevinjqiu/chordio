module github.com/kevinjqiu/chordio

go 1.13

require (
	github.com/golang/protobuf v1.3.2
	github.com/magiconair/properties v1.8.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.4.0
	go.opentelemetry.io/otel v0.4.3
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.4.3
	google.golang.org/grpc v1.27.1
)

replace go.opentelemetry.io/otel => ./vendor/opentelemetry-go/
