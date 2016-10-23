package app

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
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

	bind := GetBind(cmd)
	client, err := f.MrrClient(bind)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not create client to kubemrr: %v\n", err)
		return nil
	}

	if strings.HasPrefix(args[0], "p") {
		err = outputPods(client, f.StdOut())
	} else if strings.HasPrefix(args[0], "s") {
		err = outputServices(client, f.StdOut())
	} else {
		err = outputDeployments(client, f.StdOut())
	}

	if err != nil {
		fmt.Fprint(f.StdErr(), err)
		return nil
	}

	return nil
}

func outputPods(client MrrClient, out io.Writer) error {
	pods, err := client.Pods()
	if err != nil {
		return err
	}
	log.WithField("pods", pods).Debugf("Got pods")

	for i, pod := range pods {
		if i != 0 {
			out.Write([]byte(" "))
		}
		out.Write([]byte(pod.Name))
	}

	return nil
}

func outputServices(client MrrClient, out io.Writer) error {
	services, err := client.Services()
	if err != nil {
		return err
	}
	log.WithField("services", services).Debugf("Got services")

	for i, svc := range services {
		if i != 0 {
			out.Write([]byte(" "))
		}
		out.Write([]byte(svc.Name))
	}

	return nil
}

func outputDeployments(client MrrClient, out io.Writer) error {
	deployments, err := client.Deployments()
	if err != nil {
		return err
	}
	log.WithField("deployments", deployments).Debugf("Got deployments")

	for i, deployment := range deployments {
		if i != 0 {
			out.Write([]byte(" "))
		}
		out.Write([]byte(deployment.Name))
	}

	return nil
}

func parseKubeConfig(filename string) (Config, error) {
	res := Config{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return res, fmt.Errorf("could not read file %s: %s", filename, err)
	}

	err = yaml.Unmarshal(raw, &res)
	if err != nil {
		return res, fmt.Errorf("could not parse file %s: %s", filename, err)
	}

	return res, nil
}
