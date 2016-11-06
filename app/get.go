package app

import (
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
    - configmap, configmaps

  To filter alive resources it uses current context from the ~/.kube/conf file.
  Additionally, it accepts --namespace, --context, --server and --cluster parameters
  in "kubectl-flags".

EXAMPLE
  kubemrr -a 0.0.0.0 -p 33033 --kubect-flags="--namespace prod" get pod
`,
		Run: func(cmd *cobra.Command, args []string) {
			RunCommon(cmd)
			err := RunGet(f, cmd, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	AddCommonFlags(cmd)
	cmd.Flags().String("kubectl-flags", "", "An arbitrary string that contains flags accepted by kubectl")
	return cmd
}

func RunGet(f Factory, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		fmt.Fprintf(f.StdErr(), "You must specify the resource type")
		return nil
	}

	if len(args) > 1 {
		fmt.Fprintf(f.StdErr(), "Only one argument is expected")
		return nil
	}

	regex := "(po|pod|pods|svc|service|services|deployment|deployments|configmap|configmaps)"
	argMatcher, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not compile regular expression: %v\n", err)
		return nil
	}

	if !argMatcher.MatchString(args[0]) {
		fmt.Fprintf(f.StdErr(), "Unsupported resource type [%s]\n", args[0])
		return nil
	}

	conf, err := f.HomeKubeconfig()
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not read kubeconfig: %s\n", err)
		return nil
	}

	kubectlFlags := parseKubectlFlags(getKubectlFlags(cmd))

	bind := GetBind(cmd)
	client, err := f.MrrClient(bind)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not create client to kubemrr: %v\n", err)
		return nil
	}

	if strings.HasPrefix(args[0], "p") {
		err = outputNames(client, makeFilterFor("pod", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "s") {
		err = outputNames(client, makeFilterFor("service", &conf, kubectlFlags), f.StdOut())
	} else if strings.HasPrefix(args[0], "c") {
		err = outputNames(client, makeFilterFor("configmap", &conf, kubectlFlags), f.StdOut())
	} else {
		err = outputNames(client, makeFilterFor("deployment", &conf, kubectlFlags), f.StdOut())
	}

	if err != nil {
		fmt.Fprint(f.StdErr(), err)
		return nil
	}

	return nil
}

func getKubectlFlags(cmd *cobra.Command) string {
	r, err := cmd.Flags().GetString("kubectl-flags")
	if err != nil {
		log.Fatal(err)
	}
	return r
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

	log.WithField("in", in).WithField("out", res).Debug("Parsed kubectl flags")
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
