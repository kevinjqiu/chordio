package main

import (
	"fmt"
	"github.com/kevinjqiu/chordio"
	"github.com/kevinjqiu/chordio/cmd/client"
	"github.com/kevinjqiu/chordio/cmd/common"
	"github.com/kevinjqiu/chordio/cmd/server"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/spf13/cobra"
	"os"
)

var flushFunc telemetry.FlushFunc

func NewRootCommand() *cobra.Command {
	var flags common.CommonFlags
	cmd := &cobra.Command{
		Use:     "chordio",
		Short:   "A distributed hash table based on the Chord P2P system",
		Version: "0.1.0",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			chordio.SetLogLevel(flags.Loglevel)
			return nil
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if flushFunc != nil {
				flushFunc()
			}
			return nil
		},
	}

	common.AddCommonPflags(cmd, &flags)
	cmd.AddCommand(server.NewServerCommand())
	cmd.AddCommand(client.NewClientCommand())
	return cmd
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
