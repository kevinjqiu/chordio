package client

import (
	"github.com/kevinjqiu/chordio/cmd/common"
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
	flushFunc   telemetry.FlushFunc
)

func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "chord client commands",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			tcon, err := common.GetTelemetryConfig(cmd.Parent())
			if err != nil {
				return err
			}

			flushFunc, err = telemetry.Init("chordio/client", tcon)
			if err != nil {
				return err
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
				return err
			}
			chordClient = pb.NewChordClient(conn)
			return nil
		},
	}

	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newJoinCommand())
	cmd.AddCommand(newStabilizeCommand())
	return cmd
}
