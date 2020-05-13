package server

import (
	"fmt"
	"github.com/kevinjqiu/chordio"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/chord/node"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math"
	"strconv"
	"strings"
	"time"
)

type runFlags struct {
	id                    string
	m                     uint32
	bind                  string
	loglevel              string
	stabilizationDisabled bool
	stabilizationPeriod   time.Duration
	stabilizationJitter   time.Duration
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

			var id chord.ID
			if flags.id == "" {
				id = node.AssignID([]byte(bind), chord.Rank(flags.m))
			} else {
				uintID, err := strconv.ParseUint(flags.id, 10, 64)
				if err != nil {
					return errors.Wrap(err, "cannot parse id")
				}
				if float64(id) >= math.Pow(2.0, float64(flags.m)) {
					return errors.New("invalid id: id must between 0 and 2**m")
				}
				id = chord.ID(uintID)
			}

			flushFunc, err := telemetry.Init(fmt.Sprintf("chordio/#%d", id), telemetry.Config{})
			defer flushFunc()

			config := chordio.Config{
				ID:   id,
				M:    chord.Rank(flags.m),
				Bind: flags.bind,
				Stabilization: chordio.StabilizationConfig{
					Disabled: flags.stabilizationDisabled,
					Period:   flags.stabilizationPeriod,
					Jitter:   flags.stabilizationJitter,
				},
			}

			server, err := chordio.NewServer(config)
			if err != nil {
				return err
			}
			if err := server.Serve(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.id, "id", "i", "", "assign an ID to the node")
	cmd.Flags().Uint32VarP(&flags.m, "rank", "m", 0, "the rank of the ring")
	cmd.Flags().StringVarP(&flags.bind, "bind", "b", "localhost:2000", "bind address")
	cmd.Flags().StringVarP(&flags.loglevel, "loglevel", "l", "info", "log level")
	cmd.Flags().BoolVarP(&flags.stabilizationDisabled, "stabilization-disabled", "d", false, "disable stabilization for debugging")
	cmd.Flags().DurationVarP(&flags.stabilizationPeriod, "stabilization-period", "p", 10*time.Second, "set the stabilization run interval")
	cmd.Flags().DurationVarP(&flags.stabilizationJitter, "stabilization-jitter", "j", 5*time.Second, "set the stabilization run jitter to avoid all nodes run stabilization at the same time")
	return cmd
}
