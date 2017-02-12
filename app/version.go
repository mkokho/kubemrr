package app

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	VERSION = "1.2.0-beta1"
)

func NewVersionCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(f.StdOut(), "kubemrr-%s\n", VERSION)
		},
	}

	return cmd
}
