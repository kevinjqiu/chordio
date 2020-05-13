package client

import (
	"context"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"time"
)

func newFixFingersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "fixfingers",
		Short:        "Run fixfingers (debugging)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer flushFunc()

			md := metadata.Pairs(
				"timestamp", time.Now().Format(time.StampNano),
				"operation", "fixfingers",
			)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			_, err := chordClient.X_FixFinger(ctx, &pb.FixFingerRequest{})
			return err
		},
	}
	return cmd
}
