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
		Short: "Asks mirror for resources",
		Long: `
Ask mirror of Kubernetes API server for resources.

Supported resources are:
  - po, pod, pod
  - svc, service, services

Currently it outputs only names separated by space.
This is enought to make autocompletion works fast.
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
	cmd.Flags().String("kubectl-command", "kubectl", "Command typed in the prompt. Use to pass kubectl arguments")
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

	regex := "(po|pod|pods|svc|service|services|deployment|deployments)"
	argMatcher, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not compile regular expression: %v\n", err)
		return nil
	}

	if !argMatcher.MatchString(args[0]) {
		fmt.Fprintf(f.StdErr(), "Unsupported resource type [%s]\n", args)
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
	r, err := cmd.Flags().GetString("kubectl-command")
	if err != nil {
		log.Fatal(err)
	}
	return r
}

type KubectlFlags struct {
	namespace string
	context   string
	cluster   string
}

var (
	namespaceFlagRegex = regexp.MustCompile(`--namespace[ =]([\w]+)`)
)

func parseKubectlFlags(in string) *KubectlFlags {
	res := KubectlFlags{}
	namespaces := namespaceFlagRegex.FindAllStringSubmatch(in, -1)
	for i := range namespaces {
		res.namespace = namespaces[i][1]
	}
	log.WithField("in", in).WithField("out", res).Debug("Parsed kubectl flags")
	return &res
}

func makeFilterFor(kind string, conf *Config, flags *KubectlFlags) MrrFilter {
	f := MrrFilter{}
	if conf != nil {
		f = conf.makeFilter()
	}
	if flags != nil {
		f.Namespace = flags.namespace
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
