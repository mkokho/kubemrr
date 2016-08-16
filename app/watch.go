package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/url"
	"time"
)

type (
	Filter struct {
		NamePrefix string
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

	c := NewMrrCache()
	kc := NewKubeClient()
	kc.BaseURL = url
	go loopUpdate(c, kc)
	err = ServeMrrCache(l, c)
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

func loopUpdate(c *MrrCache, kc *KubeClient) {
	pods, err := kc.getPods()
	if err != nil {
		log.Printf("Could not get pods from %v: %v", kc.BaseURL, err)
	}

	if pods != nil {
		log.Printf("Received %d pods from %v", len(pods), kc.BaseURL)
		c.setPods(pods)
	}
	time.Sleep(time.Millisecond * 500)
	loopUpdate(c, kc)
}
