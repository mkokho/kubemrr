package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mkokho/kubemrr/pkg"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch one or several Kubernetes API servers",
	Run: func(cmd *cobra.Command, args []string) {
		pkg.RunWatch(cmd)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
	pkg.AddCommonFlags(watchCmd)
}
