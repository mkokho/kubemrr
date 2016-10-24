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
		for _, w := range []string{"WatchPods", "WatchServices", "WatchDeployments"} {
			if kc.hits[w] < 1 {
				t.Errorf("Not enough %s requests for server %s", w, s)
			}
		}
	}

	if c.pods == nil {
		t.Errorf("Pods in the cache has not been updated")
	}

	if c.services == nil {
		t.Errorf("Services in the cache has not been updated")
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
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "a"}}},
		{Deleted, &Pod{ObjectMeta: ObjectMeta{Name: "a"}}},
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "pod1"}}},
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "pod0"}}},
		{Modified, &Pod{ObjectMeta: ObjectMeta{Name: "pod1", ResourceVersion: "v2"}}},
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "z"}}},
		{Deleted, &Pod{ObjectMeta: ObjectMeta{Name: "z"}}},
	}

	loopWatchPods(c, kc)
	time.Sleep(10 * time.Millisecond)

	//order is matter is slice
	expected := []KubeObject{kc.podEvents[4].Pod.untype(), kc.podEvents[3].Pod.untype()}
	actual := c.objects[kc.Server()]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Cache version %+v is not equal to expected %+v", actual, expected)
	}
}

func TestLoopWatchPodsWithNamespaces(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.podEvents = []*PodEvent{
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "a", Namespace: "y1"}}},
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "a", Namespace: "y2"}}},
		{Added, &Pod{ObjectMeta: ObjectMeta{Name: "a", Namespace: "y3"}}},
		{Deleted, &Pod{ObjectMeta: ObjectMeta{Name: "a", Namespace: "y3"}}},
	}

	loopWatchPods(c, kc)
	time.Sleep(10 * time.Millisecond)

	//order is matter is slice
	expected := []KubeObject{kc.podEvents[0].Pod.untype(), kc.podEvents[1].Pod.untype()}
	actual := c.objects[kc.Server()]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Cache version %+v is not equal to expected %+v", actual, expected)
	}
}

func TestLoopWatchServicesFailure(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.errors["WatchServices"] = errors.New("Test Error")

	loopWatchServices(c, kc)

	time.Sleep(10 * time.Millisecond)
	if kc.hits["WatchServices"] < 2 {
		t.Errorf("Not enough WatchServices calls")
	}
}

func TestLoopWatchServices(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.serviceEvents = []*ServiceEvent{
		{Added, &Service{ObjectMeta: ObjectMeta{Name: "service0"}}},
		{Added, &Service{ObjectMeta: ObjectMeta{Name: "service1"}}},
		{Modified, &Service{ObjectMeta: ObjectMeta{Name: "service1", ResourceVersion: "v2"}}},
		{Deleted, &Service{ObjectMeta: ObjectMeta{Name: "service0"}}},
	}

	loopWatchServices(c, kc)
	time.Sleep(10 * time.Millisecond)

	expected := *kc.serviceEvents[2].Service
	if !reflect.DeepEqual(*c.services["service1"], expected) {
		t.Errorf("Cache version %v is not equal to expected %v", c.services["service1"], expected)
	}

	if _, ok := c.services["service0"]; ok {
		t.Errorf("Pod [%s] should have been deleted", "service0")
	}
}

func TestLoopWatchDeploymentsFailure(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.errors["WatchDeployments"] = errors.New("Test Error")

	loopWatchDeployments(c, kc)

	time.Sleep(10 * time.Millisecond)
	if kc.hits["WatchDeployments"] < 2 {
		t.Errorf("Not enough WatchDeployments calls")
	}
}

func TestLoopWatchDeployments(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kc.deploymentEvents = []*DeploymentEvent{
		{Added, &Deployment{ObjectMeta: ObjectMeta{Name: "deployment0"}}},
		{Added, &Deployment{ObjectMeta: ObjectMeta{Name: "deployment1"}}},
		{Modified, &Deployment{ObjectMeta: ObjectMeta{Name: "deployment1", ResourceVersion: "v2"}}},
		{Deleted, &Deployment{ObjectMeta: ObjectMeta{Name: "deployment0"}}},
	}

	loopWatchDeployments(c, kc)
	time.Sleep(10 * time.Millisecond)

	expected := *kc.deploymentEvents[2].Deployment
	if !reflect.DeepEqual(*c.deployments["deployment1"], expected) {
		t.Errorf("Cache version %v is not equal to expected %v", c.deployments["deployment1"], expected)
	}

	if _, ok := c.deployments["deployment0"]; ok {
		t.Errorf("Pod [%s] should have been deleted", "deployment0")
	}
}
