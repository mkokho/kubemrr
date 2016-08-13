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

//args := &server.Args{7,8}
//var reply int
//err = client.Call("Arith.Multiply", args, &reply)
// client.Call("Resources.Pods",
func Server(port int) int {
	c := Cache {}
	rpc.Register(&c)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if e != nil {
		log.Printf("Watcher has failed to start: %v", e)
		return 1
	}
	go http.Serve(l, nil)

	log.Printf("Watcher has been started on port %v\n", port)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
		case <-term:
			log.Println("Received SIGTERM, exiting gracefully...")
	}

	log.Println("Watcher has been stopped")
	return 0
}

func (c *Cache) Pods(f *Filter, pods *[]Pod) error {
	log.Printf("Received request for pods")
	*pods = c.getPods()

	return nil
}
