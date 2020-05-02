package client

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type joinFlags struct {
	introducerURL string
}

func newJoinCommand() *cobra.Command {
	var flags joinFlags
	cmd := &cobra.Command{
		Use:   "join",
		Short: "join node with another",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.introducerURL == "" {
				return errors.New("--introducer-url must be set")
			}

			joinReq := pb.JoinRingRequest{
				Introducer: &pb.Node{
					Bind: flags.introducerURL,
				},
			}
			resp, err := chordClient.JoinRing(context.Background(), &joinReq)
			if err != nil {
				logrus.Fatal(err)
			}

			fmt.Println(resp)
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.introducerURL, "introducer-url", "i", "", "introducer's address")
	return cmd
}
