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
	f := &TestFactory{stdErr: buf}
	cmd := NewWatchCommand(f)

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
			args: []string{"first", "second"},
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
	url, _ := url.Parse("test-url")
	kc := &TestKubeClient{baseURL: url}
	c := NewMrrCache()
	f := &TestFactory{kubeClient: kc, mrrCache: c}
	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.Flags().Set("interval", "4ms")
	go cmd.Run(cmd, []string{"http://k8s-server.example.org"})

	time.Sleep(10 * time.Millisecond)
	if kc.hitsGetPods < 2 {
		t.Errorf("Not enough GetPods requests")
	}

	if c.pods == nil {
		t.Errorf("Pods in the cache has not been updated")
	}

	if kc.hitsGetServices < 2 {
		t.Errorf("Not enough GetService requests")
	}

	if c.services == nil {
		t.Errorf("Services in the cache has not been updated")
	}

	if kc.hitsGetDeployments < 2 {
		t.Errorf("Not enough Getdeployments requests")
	}

	if c.deployments == nil {
		t.Errorf("Deployments in the cache has not been updated")
	}
}

func TestLoopWatchPodsFailure(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.errors["WatchPods"] = errors.New("Test Error")

	loopWatchPods(c, kc)

	time.Sleep(10 * time.Millisecond)
	if kc.hits["WatchPods"] < 2 {
		t.Errorf("Not enough WatchPods calls")
	}
}

func TestLoopWatchPods(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.podEvents = []*PodEvent{
		{Added, Pod{ObjectMeta: ObjectMeta{Name: "pod0"}}},
		{Added, Pod{ObjectMeta: ObjectMeta{Name: "pod1"}}},
		{Modified, Pod{ObjectMeta: ObjectMeta{Name: "pod1", ResourceVersion: "v2"}}},
		{Deleted, Pod{ObjectMeta: ObjectMeta{Name: "pod0"}}},
	}

	loopWatchPods(c, kc)
	time.Sleep(10 * time.Millisecond)

	expected := kc.podEvents[2].Pod
	if !reflect.DeepEqual(*c.pods["pod1"], expected) {
		t.Errorf("Cache version %v is not equal to expected %v", c.pods["pod1"], expected)
	}

	if _, ok := c.pods["pod2"]; ok {
		t.Errorf("Pod [%s] should have been deleted", "pod2")
	}
}
