package app

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/url"
)

func NewWatchCommand(f Factory) *cobra.Command {
	var watchCmd = &cobra.Command{
		Use:   "watch [flags] [url]",
		Short: "Starts a mirror of Kubernetes API servers",
		Long: `
Starts a mirror of Kubernetes API servers
`,
		Run: func(cmd *cobra.Command, args []string) {
			RunCommon(cmd)
			RunWatch(f, cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	return watchCmd
}

func RunWatch(f Factory, cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Fprint(f.StdErr(), "No URL given")
		return
	}

	bind := GetBind(cmd)

	l, err := net.Listen("tcp", bind)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror failed to bind on %s: %v", bind, err)
		return
	}

	c := f.MrrCache()

	for i := range args {
		url, err := url.Parse(args[i])
		if err != nil || url.Scheme == "" {
			fmt.Fprintf(f.StdErr(), "Could not parse [%s] as URL: %v", args[i], err)
			return
		}

		kc := f.KubeClient(url)
		log.Infof("Created kube client for %s", args[i])

		loopWatchPods(c, kc)
		loopWatchServices(c, kc)
		loopWatchDeployments(c, kc)
	}

	log.Infof("Kube Mirror is listening on %s", bind)
	err = f.Serve(l, c)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror encounered unexpected error: %v", err)
		return
	}
	log.Println("Kube Mirror has stopped")
}

func loopWatchPods(c *MrrCache, kc KubeClient) {
	events := make(chan *ObjectEvent)

	watch := func() {
		for {
			log.Infof("Started to watch pods for %s", kc.BaseURL().String())
			err := kc.WatchObjects("pod", events)
			if err != nil {
				log.Infof("Disruption while watching pods: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Infof("Received event [%s] for pod [%s]", e.Type, e.Object.Name)
				switch e.Type {
				case Deleted:
					c.deleteKubeObject(kc.Server(), *e.Object)
				case Added, Modified:
					c.updateKubeObject(kc.Server(), *e.Object)
				}
				//log.WithField("pods", c.objects).Debugf("Cached objects")
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
			log.Infof("Started to watch services for %s", kc.BaseURL().String())
			err := kc.WatchServices(events)
			if err != nil {
				log.Infof("Disruption while watching services: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Infof("Received event [%s] for service [%s]", e.Type, e.Service.Name)
				switch e.Type {
				case Deleted:
					c.removeService(e.Service)
				case Added, Modified:
					c.updateService(e.Service)
				}
				log.WithField("services", c.services).Debugf("Cached services")
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
			log.Infof("Started to watch deployments for %s", kc.BaseURL().String())
			err := kc.WatchDeployments(events)
			if err != nil {
				log.Infof("Disruption while watching services: %s", err)
			}
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				log.Infof("Received event [%s] for deployment [%s]", e.Type, e.Deployment.Name)
				switch e.Type {
				case Deleted:
					c.removeDeployment(e.Deployment)
				case Added, Modified:
					c.updateDeployment(e.Deployment)
				}
				log.WithField("deployments", c.deployments).Debugf("Cached deployments")
			}
		}
	}

	go watch()
	go update()
}
