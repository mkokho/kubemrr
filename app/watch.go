package app

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/url"
	"strings"
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

  Mirrored resources: pods, services, deployments, configmaps, namespaces

  By default, "get pod" returns pods from all servers and all namespaces.
  See help for "get" command to know how to filter.

EXAMPLE:
  kubemrr -a 0.0.0.0 -p 33033 watch https://kube-api-1.com https://kube-api-2.com
  kubemrr -a 0.0.0.0 -p 33033 get pod

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunCommon(cmd); err != nil {
				return err
			}
			return RunWatch(f, cmd, args)
		},
	}

	AddCommonFlags(watchCmd)
	watchCmd.Flags().Duration("interval", 2*time.Minute, "Interval between requests to the server")
	watchCmd.Flags().String("only", "", "Coma-separated names of resources to watch, empty to watch all supported")
	return watchCmd
}

func RunWatch(f Factory, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("no URL given")
	}

	bind, err := GetBind(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	l, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("failed to bind on %s: %v", bind, err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return errors.New("could not parse value of --interval")
	}

	enabledResources, err := cmd.Flags().GetString("only")
	if err != nil {
		return errors.New("could not parse value of --only")
	}

	c := f.MrrCache()

	for i := range args {
		url, err := url.Parse(args[i])
		if err != nil || url.Scheme == "" {
			return fmt.Errorf("could not parse [%s] as URL: %v", args[i], err)
		}

		kc := f.KubeClient(url)
		log.WithField("server", kc.Server().URL).Info("created client")

		for _, k := range []string{"pod"} {
			if isWatching(k, enabledResources) {
				loopWatchObjects(c, kc, k)
			}
		}

		for _, k := range []string{"service", "deployment", "configmap", "namespace"} {
			if isWatching(k, enabledResources) {
				loopGetObjects(c, kc, k, interval)
			}
		}
	}

	log.WithField("bind", bind).Info("started to listen")
	err = f.Serve(l, c)
	if err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}

	return errors.New("kubemrr has stopped")
}

func isWatching(r string, rs string) bool {
	return len(rs) == 0 || strings.Contains(rs, r)
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
			l.Info("updating objects")
			objects, err := kc.GetObjects(kind)
			if err != nil {
				l.WithField("error", err).Error("unexpected error while updating objects")
				time.Sleep(10 * time.Second)
				continue
			}

			l.WithField("objects", objects).Debug("received objects")
			c.deleteKubeObjects(kc.Server(), kind)
			for i := range objects {
				c.updateKubeObject(kc.Server(), objects[i])
			}
			l.Infof("put %d objects into cache", len(objects))

			time.Sleep(interval)
		}
	}

	go update()
}
