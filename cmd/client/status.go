package client

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"time"
)

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "status of the chord server",
		Run: func(cmd *cobra.Command, args []string) {
			md := metadata.Pairs(
				"timestamp", time.Now().Format(time.StampNano),
				"operation", "status",
			)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			resp, err := chordClient.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println("NodeID:", resp.Node.GetId())
			fmt.Println("Addr:", resp.Node.GetBind())
			fmt.Println("Pred:", resp.Node.GetPred())
			fmt.Println("Succ:", resp.Node.GetSucc())
		},
	}
	return cmd
}
