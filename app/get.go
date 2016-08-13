package app

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/rpc"
	"os"
)

func NewGetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Asks mirror for resources",
		Long:  `Asks mirror for resources`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGet(cmd, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	AddCommonFlags(cmd)
	return cmd
}

func RunGet(cmd *cobra.Command, out io.Writer) (err error) {
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

	for _, pod := range pods {
		out.Write([]byte(pod.Name))
	}

	return nil
}
