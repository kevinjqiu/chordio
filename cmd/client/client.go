package client

import (
	"github.com/kevinjqiu/chordio"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"os"
)

var (
	chordClient pb.ChordClient
	loglevel string
)

func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "client",
		Short: "chord client commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			chordioURL := os.Getenv("CHORDIO_URL")
			if chordioURL == "" {
				logrus.Fatal("CHORDIO_URL environment variable must be set")
			}
			conn, err := grpc.Dial(chordioURL, grpc.WithInsecure())
			if err != nil {
				logrus.Fatal(err)
			}
			chordClient = pb.NewChordClient(conn)
			chordio.SetLogLevel(loglevel)
		},
	}

	cmd.PersistentFlags().StringVarP(&loglevel, "loglevel", "l", "info", "log level")
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newJoinCommand())
	return cmd
}
