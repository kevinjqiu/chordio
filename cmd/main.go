package main

import (
	"fmt"
	"github.com/kevinjqiu/chordio/cmd/client"
	"github.com/kevinjqiu/chordio/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chordio",
		Short:   "A distributed hash table based on the Chord P2P system",
		Version: "0.1.0",
	}

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
