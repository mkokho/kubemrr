package app

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/rpc"
	"os"
	"strings"
)

func NewGetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Asks mirror for resources",
		Long:  `Asks mirror for resources`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGet(cmd, args, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	AddCommonFlags(cmd)
	return cmd
}

func RunGet(cmd *cobra.Command, args []string, out io.Writer) (err error) {
	if args[0] != "pod" {
		return nil
	}

	bind := GetBind(cmd)
	client, err := rpc.DialHTTP("tcp", bind)
	if err != nil {
		return
	}

	f := Filter{}
	var pods []Pod
	err = client.Call("Cache.Pods", f, &pods)
	if err != nil {
		return
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
