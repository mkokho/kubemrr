package app

import (
	"net"
	"net/rpc"
	"reflect"
	"testing"
)

func TestCallPods(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	c := NewCache()
	c.pods = []Pod{
		Pod{ObjectMeta: ObjectMeta{Name: "pod1"}},
		Pod{ObjectMeta: ObjectMeta{Name: "pod2"}},
	}
	go Serve(l, c)

	conn, err := rpc.DialHTTP("tcp", l.Addr())
	if err != nil {
		return
	}

	var pods []Pod
	err = conn.Call("Cache.Pods", nil, &pods)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(pods, c.pods) {
		t.Errorf("Expected pods %v, found %v", c.pods, pods)
	}
}
