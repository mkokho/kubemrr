package pkg

import (
	"os/signal"
	"os"
	"syscall"
	"log"
	"net/rpc"
	"net"
	"net/http"
	"github.com/spf13/cobra"
)

type Filter struct {
	NamePrefix string
}

func RunWatch(cmd *cobra.Command) {
	bind := GetBind(cmd)
	c := Cache {}
	rpc.Register(&c)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("Kube Mirror failed to start on %s: %v", bind, err)
	}
	go http.Serve(l, nil)
	log.Printf("Kube Mirror started on %s\n", bind)
	waitForSigterm()
	log.Println("Kube Mirror stopped")
}

func waitForSigterm() {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
		case <-term:
			log.Println("Received SIGTERM, exiting gracefully...")
	}
}

func (c *Cache) Pods(f *Filter, pods *[]Pod) error {
	log.Printf("Received request for pods")
	*pods = c.getPods()

	return nil
}
