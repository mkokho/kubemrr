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
		Short: "Start a mirror of Kubernetes API servers",
		Long: `
DESCRIPTION:
  Start a mirror of Kubernetes API servers.

  It maintans several connections to each of the given API servers.
  On each connection it will listen for changes happened in the Kubernetes cluster.
  The names of the alive resources are available by "get" command.

  Mirrored resources: pods, services. deployments.

  By default, "get pod" returns pods from all servers and all namespaces.
  See help for "get" command to know how to filter.

EXAMPLE:
  kubemrr -a 0.0.0.0 -p 33033 watch https://kube-api-1.com https://kube-api-2.com
  kubemrr -a 0.0.0.0 -p 33033 get pod

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

		loopWatchObjects(c, kc, "pod")
		loopWatchObjects(c, kc, "service")
		loopWatchObjects(c, kc, "deployment")
	}

	log.Infof("Kube Mirror is listening on %s", bind)
	err = f.Serve(l, c)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Kube Mirror encounered unexpected error: %v", err)
		return
	}
	log.Println("Kube Mirror has stopped")
}

func loopWatchObjects(c *MrrCache, kc KubeClient, kind string) {
	events := make(chan *ObjectEvent)
	l := log.WithField("kind", kind).WithField("server", kc.Server().URL)

	watch := func() {
		for {
			l.Info("started to watch")
			err := kc.WatchObjects(kind, events)
			fields := log.Fields{}
			if err != nil {
				fields["error"] = err.Error()
			}
			l.WithFields(fields).Info("watch connection was closed, retrying")
			c.deleteKubeObjects(kc.Server(), kind)
		}
	}

	update := func() {
		for {
			select {
			case e := <-events:
				l.
					WithField("name", e.Object.Name).
					WithField("type", e.Type).
					Info("received event")
				switch e.Type {
				case Deleted:
					c.deleteKubeObject(kc.Server(), *e.Object)
				case Added, Modified:
					c.updateKubeObject(kc.Server(), *e.Object)
				}
				l.WithField("cache", c.objects).Debugf("objects in cache")
			}
		}
	}

	go watch()
	go update()
}
