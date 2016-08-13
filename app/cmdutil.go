package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address where mirror accessible")
	cmd.Flags().IntP("port", "p", 33033, "The port on mirror is accessible")
}

func GetBind(cmd *cobra.Command) string {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		log.Fatal(err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s:%d", address, port)
}
