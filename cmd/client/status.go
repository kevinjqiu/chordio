package client

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"strconv"
	"time"
)

func printFT(ft *pb.FingerTable, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	writer := tablewriter.NewWriter(w)
	writer.SetHeader([]string{"Start", "[Start, End)", "Successor Node #"})
	for _, fte := range ft.Entries {
		writer.Append([]string{
			strconv.Itoa(int(fte.Start)),
			fmt.Sprintf("[%d, %d)", fte.Start, fte.End),
			strconv.Itoa(int(fte.NodeID)),
		})
	}
	writer.Render()
}

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

			resp, err := chordClient.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{
				IncludeFingerTable: true,
			})
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println("NodeID:", resp.Node.GetId())
			fmt.Println("Addr:", resp.Node.GetBind())
			fmt.Println("Pred:", resp.Node.GetPred())
			fmt.Println("Succ:", resp.Node.GetSucc())
			printFT(resp.Ft, nil)
		},
	}
	return cmd
}
