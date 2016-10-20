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
	watchCmd.Flags().Duration("interval", 500*time.Millisecond, "Interval between requests to the server")
	return watchCmd
}

func RunWatch(f Factory, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(f.StdErr(), "You must specify URL of Kubernetes API")
		return
	}

	//TODO remove unused interval flag
	/*	interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			fmt.Fprintf(f.StdErr(), "Could not parse value of --interval")
			return
		}*/

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
	loopWatchPods(c, kc)
	loopWatchServices(c, kc)
	loopWatchDeployments(c, kc)
	err = f.Serve(l, c)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror encounered unexpected error: %v", err)
		return
	}

	log.Println("Kube Mirror has stopped")
}

func loopWatchPods(c *MrrCache, kc KubeClient) {
	events := make(chan *PodEvent)

	watch := func() {
		for {
			log.Printf("Started to watch pods")
			err := kc.WatchPods(events)
			if err != nil {
				log.Printf("Disruption while watching pods: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Printf("Received event [%s] for pod [%s]\n", e.Type, e.Pod.Name)
				switch e.Type {
				case Deleted:
					c.removePod(e.Pod)
				case Added, Modified:
					c.updatePod(e.Pod)
				}
			}
		}
	}

	go watch()
	go update()
}

func loopWatchServices(c *MrrCache, kc KubeClient) {
	events := make(chan *ServiceEvent)

	watch := func() {
		for {
			log.Printf("Started to watch services")
			err := kc.WatchServices(events)
			if err != nil {
				log.Printf("Disruption while watching services: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Printf("Received event [%s] for service [%s]\n", e.Type, e.Service.Name)
				switch e.Type {
				case Deleted:
					c.removeService(e.Service)
				case Added, Modified:
					c.updateService(e.Service)
				}
			}
		}
	}

	go watch()
	go update()
}

func loopWatchDeployments(c *MrrCache, kc KubeClient) {
	events := make(chan *DeploymentEvent)

	watch := func() {
		for {
			log.Printf("Started to watch deployments")
			err := kc.WatchDeployments(events)
			if err != nil {
				log.Printf("Disruption while watching services: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Printf("Received event [%s] for deployment [%s]\n", e.Type, e.Deployment.Name)
				switch e.Type {
				case Deleted:
					c.removeDeployment(e.Deployment)
				case Added, Modified:
					c.updateDeployment(e.Deployment)
				}
			}
		}
	}

	go watch()
	go update()
}

func loopUpdatePods(c *MrrCache, kc KubeClient, interval time.Duration) {
	pods, err := kc.GetPods()
	if err != nil {
		log.Printf("Could not get pods from %v: %v", kc.BaseURL(), err)
	}

	if pods != nil {
		log.Printf("Received %d pods from %v", len(pods), kc.BaseURL())
		c.setPods(pods)
	}
	time.Sleep(interval)
	loopUpdatePods(c, kc, interval)
}

func loopUpdateServices(c *MrrCache, kc KubeClient, interval time.Duration) {
	services, err := kc.GetServices()
	if err != nil {
		log.Printf("Could not get services from %v: %v", kc.BaseURL(), err)
	}

	if services != nil {
		log.Printf("Received %d services from %v", len(services), kc.BaseURL())
		c.setServices(services)
	}
	time.Sleep(interval)
	loopUpdateServices(c, kc, interval)
}

func loopUpdateDeployments(c *MrrCache, kc KubeClient, interval time.Duration) {
	deployments, err := kc.GetDeployments()
	if err != nil {
		log.Printf("Could not get deployments from %v: %v", kc.BaseURL(), err)
	}

	if deployments != nil {
		log.Printf("Received %d deployments from %v", len(deployments), kc.BaseURL())
		c.setDeployments(deployments)
	}
	time.Sleep(interval)
	loopUpdateDeployments(c, kc, interval)
}
