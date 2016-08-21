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

func NewWatchCommand(f Factory) *cobra.Command {
	var watchCmd = &cobra.Command{
		Use:   "watch [flags] [url]",
		Short: "Starts a mirror of one Kubernetes API server",
		Long: `
Starts a mirror of one Kubernetes API server
`,
		Run: func(cmd *cobra.Command, args []string) {
			RunWatch(f, cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	return watchCmd
}

func RunWatch(f Factory, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(f.StdErr(), "You must specify URL of Kubernetes API")
		return
	}

	bind := GetBind(cmd)

	l, err := net.Listen("tcp", bind)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror failed to bind on %s: %v", bind, err)
		return
	}

	url, err := url.ParseRequestURI(args[0])
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not parse [%s] as URL: %v", args[0], err)
		return
	}

	log.Printf("Kube Mirror is listening on %s\n", bind)

	c := f.MrrCache()
	kc := f.KubeClient(url)
	go loopUpdatePods(c, kc)
	go loopUpdateServices(c, kc)
	err = f.Serve(l, c)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror encounered unexpected error: %v", err)
		return
	}

	log.Println("Kube Mirror has stopped")
}

func loopUpdatePods(c *MrrCache, kc KubeClient) {
	pods, err := kc.GetPods()
	if err != nil {
		log.Printf("Could not get pods from %v: %v", kc.BaseURL(), err)
	}

	if pods != nil {
		log.Printf("Received %d pods from %v", len(pods), kc.BaseURL())
		c.setPods(pods)
	}
	time.Sleep(time.Millisecond * 500)
	loopUpdatePods(c, kc)
}

func loopUpdateServices(c *MrrCache, kc KubeClient) {
	services, err := kc.GetServices()
	if err != nil {
		log.Printf("Could not get services from %v: %v", kc.BaseURL, err)
	}

	if services != nil {
		log.Printf("Received %d services from %v", len(services), kc.BaseURL)
		c.setServices(services)
	}
	time.Sleep(time.Millisecond * 500)
	loopUpdateServices(c, kc)
}
