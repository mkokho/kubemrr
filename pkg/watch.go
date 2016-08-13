package pkg

import (
	"os/signal"
	"os"
	"syscall"
	"log"
	"net/rpc"
	"net"
	"fmt"
	"net/http"
)

type Filter struct {
	NamePrefix string
}

func RunWatch(ro *RootOptions) {
	c := Cache {}
	rpc.Register(&c)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ro.Address, ro.Port))
	if err != nil {
		log.Fatalf("Kube Mirror failed to start: %v", err)
	}
	go http.Serve(l, nil)
	log.Printf("Kube Mirror started on %s:%v\n", ro.Address, ro.Port)
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
