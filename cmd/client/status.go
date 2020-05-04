package client

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "status of the chord server",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := chordClient.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{})
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
