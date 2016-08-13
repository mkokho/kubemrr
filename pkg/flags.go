package pkg

import (
	"github.com/spf13/cobra"
	"log"
)

type RootOptions struct {
  Port int
	Address string
}

func MustRootOptions(cmd *cobra.Command) *RootOptions {
	port, err := cmd.PersistentFlags().GetInt("port")
	if err != nil {
		log.Fatal(err)
	}

	address, err := cmd.PersistentFlags().GetString("address")
	if err != nil {
		log.Fatal(err)
	}

	return &RootOptions{
		Port: port,
		Address: address,
	}
}
