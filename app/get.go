package app

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"regexp"
	"strings"
)

func NewGetCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get [flags] [resource]",
		Short: "Ask mirror for resources",
		Long: `
DESCRIPTION:
  Ask "kubemrr watch" process for the names of alive resources

  Supported resources are:
    - po, pod, pod
    - svc, service, services
    - deployment, deployments
    - ns, namespace, namespaces
    - configmap, configmaps

  To filter alive resources it uses current context from the ~/.kube/conf file.
  Additionally, it accepts --namespace, --context, --server and --cluster parameters
  in "kubectl-flags".

EXAMPLE
  kubemrr -a 0.0.0.0 -p 33033 --kubect-flags="--namespace prod" get pod
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunCommon(cmd); err != nil {
				return err
			}
			return RunGet(f, cmd, args)
		},
	}

	AddCommonFlags(cmd)
	cmd.Flags().String("kubectl-flags", "", "An arbitrary string that contains flags accepted by kubectl")
	return cmd
}

func RunGet(f Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("no resource type is given")
	}

	if len(args) > 1 {
		return errors.New("only one argument is expected")
	}

	regex := "(po|pod|pods|svc|service|services|deployment|deployments|ns|namespace|namespaces|configmap|configmaps|no|node|nodes)"
	argMatcher, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	if !argMatcher.MatchString(args[0]) {
		return fmt.Errorf("unsupported resource type: %s", args[0])
	}

	conf, err := f.HomeKubeconfig()
	if err != nil {
		return fmt.Errorf("could not read kubeconfig: %s", err)
	}

	rawKubectlFlags, err := cmd.Flags().GetString("kubectl-flags")
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}
	kubectlFlags := parseKubectlFlags(rawKubectlFlags)

	bind, err := GetBind(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	client, err := f.MrrClient(bind)
	if err != nil {
		return fmt.Errorf("could not create client to kubemrr: %s", err)
	}

	if strings.HasPrefix(args[0], "p") {
		err = outputNames(client, makeFilterFor("pod", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "s") {
		err = outputNames(client, makeFilterFor("service", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "c") {
		err = outputNames(client, makeFilterFor("configmap", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "na") || args[0] == "ns" {
		err = outputNames(client, makeFilterFor("namespace", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "no") {
		err = outputNames(client, makeFilterFor("node", &conf, kubectlFlags), f.StdOut())
	} else {
		err = outputNames(client, makeFilterFor("deployment", &conf, kubectlFlags), f.StdOut())
	}

	if err != nil {
		return err
	}

	return nil
}

type KubectlFlags struct {
	namespace string
	context   string
	cluster   string
	server    string
}

var (
	namespaceFlagRegex = regexp.MustCompile(`--namespace[ =]([\S]+)`)
	serverFlagRegex    = regexp.MustCompile(`--server[ =]([\S]+)`)
	contextFlagRegex   = regexp.MustCompile(`--context[ =]([\S]+)`)
	clusterFlagRegex   = regexp.MustCompile(`--cluster[ =]([\S]+)`)
)

func parseKubectlFlags(in string) *KubectlFlags {
	res := KubectlFlags{}

	for _, matches := range namespaceFlagRegex.FindAllStringSubmatch(in, -1) {
		res.namespace = matches[1]
	}

	for _, matches := range serverFlagRegex.FindAllStringSubmatch(in, -1) {
		res.server = matches[1]
	}

	for _, matches := range contextFlagRegex.FindAllStringSubmatch(in, -1) {
		res.context = matches[1]
	}

	for _, matches := range clusterFlagRegex.FindAllStringSubmatch(in, -1) {
		res.cluster = matches[1]
	}

	log.WithField("in", in).WithField("out", res).Debug("parsed kubectl flags")
	return &res
}

func makeFilterFor(kind string, conf *Config, flags *KubectlFlags) MrrFilter {
	f := MrrFilter{}
	if conf != nil {
		if flags != nil && flags.context != "" {
			conf.CurrentContext = flags.context
		}
		f = conf.makeFilter()
	}
	if flags != nil {
		if flags.namespace != "" {
			f.Namespace = flags.namespace
		}
		if flags.cluster != "" {
			f.Server = conf.getCluster(flags.cluster).Server
		}
		if flags.server != "" {
			f.Server = flags.server
		}
	}
	f.Kind = kind

	if kind == "node" {
		f.Namespace = ""
	}

	return f
}

func outputNames(c MrrClient, f MrrFilter, out io.Writer) error {
	objects, err := c.Objects(f)
	if err != nil {
		return err
	}
	log.
		WithField("filter", f).
		WithField("objects", objects).
		Debugf("got objects")

	for i, o := range objects {
		if i != 0 {
			out.Write([]byte(" "))
		}
		out.Write([]byte(o.Name))
	}

	return nil
}
