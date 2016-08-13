package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kubemrr",
	Short: "kubemrr mirrors description of Kubernetes resources",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("address", "a", "127.0.0.1", "The IP address where mirror accessible")
	RootCmd.PersistentFlags().IntP("port", "p", 33033, "The port on mirror is accessible")
}


