package client

import (
	"context"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"time"
)

func newStabilizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stabilize",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer flushFunc()

			md := metadata.Pairs(
				"timestamp", time.Now().Format(time.StampNano),
				"operation", "stabilize",
			)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			_, err := chordClient.X_Stabilize(ctx, &pb.StabilizeRequest{})
			return err
		},
	}
	return cmd
}

