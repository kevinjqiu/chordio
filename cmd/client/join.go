package client

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"time"
)

type joinFlags struct {
	introducerURL string
}

func newJoinCommand() *cobra.Command {
	var flags joinFlags
	cmd := &cobra.Command{
		Use:          "join",
		Short:        "join node with another",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer flushFunc()

			if flags.introducerURL == "" {
				return errors.New("--introducer-url must be set")
			}

			joinReq := pb.JoinRingRequest{
				Introducer: &pb.Node{
					Bind: flags.introducerURL,
				},
			}

			md := metadata.Pairs(
				"timestamp", time.Now().Format(time.StampNano),
				"operation", "join",
			)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			resp, err := chordClient.JoinRing(ctx, &joinReq)
			if err != nil {
				return err
			}

			fmt.Println(resp)
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.introducerURL, "introducer-url", "i", "", "introducer's address")
	return cmd
}
