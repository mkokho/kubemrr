package app

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	VERSION = "0.6.0"
)

func NewVersionCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Prints version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(f.StdOut(), "kubemrr-%s", VERSION)
		},
	}

	return cmd
}
