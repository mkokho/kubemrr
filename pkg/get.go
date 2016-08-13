package pkg

import (
	"net/rpc"
	"io"
	"fmt"
)

func RunGet(ro *RootOptions, out io.Writer) (err error) {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", ro.Address, ro.Port))
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
