package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address where mirror is accessible")
	cmd.Flags().IntP("port", "p", 33033, "The port on which mirror is accessible")
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

type Factory interface {
	MrrClient(bind string) (MrrClient, error)
	StdOut() io.Writer
	StdErr() io.Writer
}

type DefaultFactory struct{}

func (f *DefaultFactory) MrrClient(address string) (MrrClient, error) {
	return NewMrrClient(address)
}

func (f *DefaultFactory) StdOut() io.Writer {
	return os.Stdout
}

func (f *DefaultFactory) StdErr() io.Writer {
	return os.Stderr
}
