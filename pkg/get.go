package pkg

import (
	"net/rpc"
	"io"
	"github.com/spf13/cobra"
)

func RunGet(cmd *cobra.Command, out io.Writer) (err error) {
	bind := GetBind(cmd)
	client, err := rpc.DialHTTP("tcp", bind)
	if err != nil {
		return
	}

	f := Filter {}
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
