package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"regexp"
	"strings"
)

func NewGetCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Asks mirror for resources",
		Long:  `Asks mirror for resources`,
		Run: func(cmd *cobra.Command, args []string) {
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
	if len(args) < 1 {
		fmt.Fprintf(f.StdErr(), "At least one argument is expected")
		return nil
	}

	regex := "(po|pod|pods|svc|service|services)"
	argMatcher, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not compile regular expression: %v", err)
		return nil
	}

	if !argMatcher.MatchString(args[0]) {
		fmt.Fprintf(f.StdErr(), "Expected %s, given %#v", regex, args)
		return nil
	}

	bind := GetBind(cmd)
	client, err := f.MrrClient(bind)
	if err != nil {
		fmt.Fprintf(f.StdErr(), "Could not create client to kubemrr: %v", err)
		return nil
	}

	if strings.HasPrefix(args[0], "p") {
		err = outputPods(client, f.StdOut())
	} else {
		err = outputServices(client, f.StdOut())
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

	for i, svc := range services {
		if i != 0 {
			out.Write([]byte(" "))
		}
		out.Write([]byte(svc.Name))
	}

	return nil
}
