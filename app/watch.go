package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"sync"
	"time"
)

type (
	Filter struct {
		NamePrefix string
	}

	Cache struct {
		pods []Pod
		mu   *sync.RWMutex
	}
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

func RunWatch(cmd *cobra.Command, args []string) {
	bind := GetBind(cmd)

	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("Kube Mirror failed to bind on %s: %v", bind, err)
	}

	url, err := parseArgs(args)
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}

	log.Printf("Kube Mirror is listening on %s\n", bind)

	c := NewCache()
	kc := NewKubeClient()
	kc.BaseURL = url
	go loopUpdate(c, kc)
	err = Serve(l, c)
	if err != nil {
		log.Fatalf("Kube Mirror encounered unexpected error: %v", err)
	}

	log.Println("Kube Mirror has stopped")
}

func parseArgs(args []string) (*url.URL, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Expected exactly one url as an argument")
	}

	url, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("Could not parse %s: %s", args[0], err)
	}

	return url, nil
}

func Serve(l net.Listener, cache *Cache) error {
	rpc.Register(cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}

func loopUpdate(c *Cache, kc *KubeClient) {
	pods, err := kc.getPods()
	if err != nil {
		log.Printf("Could not get pods from %v: %v", kc.BaseURL, err)
	}

	if pods != nil {
		log.Printf("Receive %d pods from %v", len(pods), kc.BaseURL)
		c.setPods(pods)
	}
	time.Sleep(time.Millisecond * 500)
	loopUpdate(c, kc)
}

func NewCache() *Cache {
	return &Cache{
		mu: &sync.RWMutex{},
	}
}

func (c *Cache) Pods(f *Filter, pods *[]Pod) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	*pods = c.pods
	return nil
}

func (c *Cache) setPods(pods []Pod) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pods = pods
}
