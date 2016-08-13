package cmd

import (

	"github.com/spf13/cobra"
	"os"

	"github.com/mkokho/kubemrr/pkg"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch one or several Kubernetes API servers",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(pkg.Server(14088))
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
