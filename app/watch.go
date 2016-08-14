package app

import (
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func NewWatchCommand() *cobra.Command {
	var watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Starts a mirror of one or several Kubernetes API servers",
		Long:  "Starts a mirror of one or several Kubernetes API servers",
		Run: func(cmd *cobra.Command, args []string) {
			RunWatch(cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	return watchCmd
}

type Filter struct {
	NamePrefix string
}

func RunWatch(cmd *cobra.Command, args []string) {
	bind := GetBind(cmd)
	c := Cache{}
	rpc.Register(&c)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("Kube Mirror failed to start on %s: %v", bind, err)
	}

	log.Printf("Kube Mirror is listening on %s\n", bind)
	http.Serve(l, nil)
	log.Println("Kube Mirror has stopped")
}

func (c *Cache) Pods(f *Filter, pods *[]Pod) error {
	log.Printf("Received request for pods")
	*pods = c.getPods()

	return nil
}
