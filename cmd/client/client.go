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
	flushFunc   telemetry.FlushFunc
)

func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "chord client commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			chordio.SetLogLevel(loglevel)

			flushFunc, err = telemetry.Init("chordio/client", telemetry.Config{})
			if err != nil {
				logrus.Fatal(err)
			}

			chordioURL := os.Getenv("CHORDIO_URL")
			if chordioURL == "" {
				logrus.Fatal("CHORDIO_URL environment variable must be set")
			}

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
	cmd.AddCommand(newStabilizeCommand())
	cmd.AddCommand(newFixFingersCommand())
	return cmd
}
