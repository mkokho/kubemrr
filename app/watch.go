package app

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/url"
	"time"
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

  Mirrored resources: pods, services, deployments, configmaps

  By default, "get pod" returns pods from all servers and all namespaces.
  See help for "get" command to know how to filter.

EXAMPLE:
  kubemrr -a 0.0.0.0 -p 33033 watch https://kube-api-1.com https://kube-api-2.com
  kubemrr -a 0.0.0.0 -p 33033 get pod

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			RunCommon(cmd)
			return RunWatch(f, cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	watchCmd.Flags().Duration("interval", 30*time.Second, "Interval between requests to the server")
	return watchCmd
}

func RunWatch(f Factory, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("no URL given")
	}

	bind := GetBind(cmd)

	l, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("failed to bind on %s: %v", bind, err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return errors.New("could not parse value of --interval")
	}

	c := f.MrrCache()

	for i := range args {
		url, err := url.Parse(args[i])
		if err != nil || url.Scheme == "" {
			return fmt.Errorf("could not parse [%s] as URL: %v", args[i], err)
		}

		kc := f.KubeClient(url)
		log.WithField("server", kc.Server().URL).Info("created client")

		loopWatchObjects(c, kc, "pod")
		loopWatchObjects(c, kc, "service")
		loopWatchObjects(c, kc, "deployment")
		loopGetObjects(c, kc, "configmap", interval)
	}

	log.WithField("bind", bind).Infof("started to listen", bind)
	err = f.Serve(l, c)
	if err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}

	return errors.New("kubemrr has stopped")
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

func loopGetObjects(c *MrrCache, kc KubeClient, kind string, interval time.Duration) {
	l := log.WithField("kind", kind).WithField("server", kc.Server().URL)
	update := func() {
		for {
			l.Info("getting objects")
			objects, err := kc.GetObjects(kind)
			if err != nil {
				l.WithField("error", err).Error("unexpected error while getting objects")
			} else {
				l.WithField("objects", objects).Debug("received objects")
			}

			for i := range objects {
				c.updateKubeObject(kc.Server(), objects[i])
			}

			time.Sleep(interval)
		}
	}

	go update()
}
