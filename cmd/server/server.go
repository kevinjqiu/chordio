package server

import (
	"github.com/kevinjqiu/chordio/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runFlags struct {
	m uint32
	bind string
}

func NewServerCommand() *cobra.Command {
	var flags runFlags
	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.m == 0 {
				return errors.New("Chord ring rank (m) must be specified")
			}

			config := pkg.Config{
				M:    pkg.Rank(flags.m),
				Bind: flags.bind,
			}

			server, err := pkg.NewServer(config)
			if err != nil {
				logrus.Fatal(err)
			}
			if err := server.Serve(); err != nil {
				logrus.Fatal(err)
			}
			return nil
		},
	}

	cmd.Flags().Uint32VarP(&flags.m, "rank", "m", 0, "the rank of the ring")
	cmd.Flags().StringVarP(&flags.bind, "bind", "b", "localhost:2000", "bind address")
	return cmd
}

