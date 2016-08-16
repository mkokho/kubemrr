package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
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

	if args[0] != "pod" && args[0] != "po" && args[0] != "pods" {
		fmt.Fprintf(stderr, "Expected one of (po|pod|pods), given %#v", args)
		return nil
	}

	bind := GetBind(cmd)
	client, err := NewMrrClient(bind)
	if err != nil {
		fmt.Fprintf(stderr, "Could not create client to kubemrr: %v", err)
		return nil
	}

	pods, err := client.Pods()
	if err != nil {
		fmt.Fprintf(stderr, "Server failed to return pods: %v", err)
		return nil
	}

	prefix := ""
	if len(args) == 2 {
		prefix = args[1]
	}

	for _, pod := range filter(pods, prefix) {
		out.Write([]byte(pod.Name))
		out.Write([]byte(" "))
	}

	return nil
}

func filter(vs []Pod, prefix string) []Pod {
	vsf := make([]Pod, 0)
	for _, v := range vs {
		if strings.HasPrefix(v.Name, prefix) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
