package app

import (
	"bytes"
	"github.com/pkg/errors"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestRunWatchInvalidArgs(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	f := NewTestFactory()
	f.stdErr = buf
	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")

	tests := []struct {
		args []string
	}{
		{
			args: []string{},
		},
		{
			args: []string{"not-a-url"},
		},
		{
			args: []string{"http://k8s-server_1.com", "not-a-url"},
		},
	}

	for i, test := range tests {
		buf.Reset()
		cmd.Run(cmd, test.args)
		if buf.Len() == 0 {
			t.Errorf("Test %d: nothing has been written to the error output, expected something", i)
		}
	}
}

func TestRunWatch(t *testing.T) {
	c := NewMrrCache()
	f := NewTestFactory()
	f.mrrCache = c

	servers := []string{
		"http://k8s-server_1.com",
		"http://k8s-server_2.com",
	}

	for _, s := range servers {
		kc := NewTestKubeClient()
		kc.baseURL, _ = url.Parse(s)
		f.kubeClients[s] = kc
	}

	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	go cmd.Run(cmd, servers)
	time.Sleep(10 * time.Millisecond)

	for s, kc := range f.kubeClients {
		for _, kind := range []string{"pod", "service", "deployment"} {
			if kc.watchObjectHits[kind] != 1 {
				t.Errorf("Unexpected number of WatchObject requests for [%s] server [%s]: %v", kind, s, kc.watchObjectHits)
			}
		}
	}
}

func TestLoopWatchObjectsFailure(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.watchObjectError = errors.New("Test Error")

	loopWatchObjects(c, kc, "o")

	time.Sleep(10 * time.Millisecond)
	if kc.watchObjectHits["o"] < 2 {
		t.Errorf("Not enough WatchObjects calls")
	}
}

func TestLoopWatchObjects(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.objectEvents = []*ObjectEvent{
		{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "a"}}},
		{Deleted, &KubeObject{ObjectMeta: ObjectMeta{Name: "a"}}},
		{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "pod1"}}},
		{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "pod0"}}},
		{Modified, &KubeObject{ObjectMeta: ObjectMeta{Name: "pod1", ResourceVersion: "v2"}}},
		{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "z"}}},
		{Deleted, &KubeObject{ObjectMeta: ObjectMeta{Name: "z"}}},
	}

	loopWatchObjects(c, kc, "o")
	time.Sleep(10 * time.Millisecond)

	//order matters in slice
	expected := []KubeObject{*kc.objectEvents[4].Object, *kc.objectEvents[3].Object}
	actual := c.objects[kc.Server()]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Cache version %+v is not equal to expected %+v", actual, expected)
	}
}
