package server

import (
	"fmt"
	"github.com/kevinjqiu/chordio"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

type runFlags struct {
	m        uint32
	bind     string
	loglevel string
}

func mustBind(bind string) string {
	var err error

	parts := strings.Split(bind, ":")
	if len(parts) != 2 {
		logrus.Fatal("Invalid bind format")
	}

	ip := parts[0]
	if ip == "" {
		ip, err = getFirstAvailableBindIP()
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		if !canBindIP(ip) {
			logrus.Fatalf("cannot bind to IP %s: %s", ip, err)
		}
	}

	return fmt.Sprintf("%s:%s", ip, parts[1])
}

func NewServerCommand() *cobra.Command {
	var flags runFlags
	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.m == 0 {
				return errors.New("Chord ring rank (m) must be specified")
			}
			chordio.SetLogLevel(flags.loglevel)

			bind := mustBind(flags.bind)

			id := chordio.AssignID([]byte(bind), chordio.Rank(flags.m))

			flushFunc, err := telemetry.Init(fmt.Sprintf("chordio/nodeid=%d", id), telemetry.Config{})
			defer flushFunc()

			config := chordio.Config{
				ID:   id,
				M:    chordio.Rank(flags.m),
				Bind: flags.bind,
			}

			server, err := chordio.NewServer(config)
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
	cmd.Flags().StringVarP(&flags.loglevel, "loglevel", "l", "info", "log level")
	return cmd
}
