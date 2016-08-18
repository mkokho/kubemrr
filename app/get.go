package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func NewGetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Asks mirror for resources",
		Long:  `Asks mirror for resources`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGet(cmd, args, os.Stdout, os.Stderr)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	AddCommonFlags(cmd)
	return cmd
}

func RunGet(cmd *cobra.Command, args []string, out io.Writer, stderr io.Writer) (err error) {
	if len(args) < 1 {
		fmt.Fprintf(stderr, "At least one argument is expected")
		return nil
	}

	regex := "(po|pod|pods|svc|service|services)"
	argMatcher, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(stderr, "Could not compile regular expression: %v", err)
		return nil
	}

	if !argMatcher.MatchString(args[0]) {
		fmt.Fprintf(stderr, "Expected %s, given %#v", regex, args)
		return nil
	}

	bind := GetBind(cmd)
	client, err := NewMrrClient(bind)
	if err != nil {
		fmt.Fprintf(stderr, "Could not create client to kubemrr: %v", err)
		return nil
	}

	if strings.HasPrefix(args[0], "p") {
		err = outputPods(client, out)
	} else {
		err = outputServices(client, out)
	}

	if err != nil {
		fmt.Fprint(stderr, err)
		return nil
	}

	return nil
}

func outputPods(client *MrrClientDefault, out io.Writer) error {
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

func outputServices(client *MrrClientDefault, out io.Writer) error {
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
