package pkg

import (
	"net/rpc"
	"io"
)

func RunGet(out io.Writer) (err error) {
	client, err := rpc.DialHTTP("tcp", "localhost:14088")
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
