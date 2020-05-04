package client

import (
	"github.com/kevinjqiu/chordio"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"os"
)

var (
	chordClient pb.ChordClient
	loglevel    string
)

func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "chord client commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			chordio.SetLogLevel(loglevel)

			chordioURL := os.Getenv("CHORDIO_URL")
			if chordioURL == "" {
				logrus.Fatal("CHORDIO_URL environment variable must be set")
			}

			flushFunc, err := telemetry.Init(telemetry.Config{})
			if err != nil {
				logrus.Fatal(err)
			}
			defer flushFunc()

			conn, err := grpc.Dial(
				chordioURL,
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer("chordio/client"))),
				grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(global.Tracer("chordio/client"))),
			)
			if err != nil {
				logrus.Fatal(err)
			}
			chordClient = pb.NewChordClient(conn)
		},
	}

	cmd.PersistentFlags().StringVarP(&loglevel, "loglevel", "l", "info", "log level")
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newJoinCommand())
	return cmd
}
