package app

import (
	"bytes"
	"net/url"
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

	if kc.hitsGetPods < 2 {
		t.Errorf("Not enough GetService requests")
	}

	if c.services == nil {
		t.Errorf("Services in the cache has not been updated")
	}

	if kc.hitsGetDeployments < 2 {
		t.Errorf("Not enough Getdeployments requests")
	}
}
