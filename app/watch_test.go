package app

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestRunWatchInvalidURLArgs(t *testing.T) {
	f := NewTestFactory()
	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.Flags().Set("kubeconfig", "test_data/kubeconfig_valid")

	tests := []struct {
		args []string
	}{
		{
			args: []string{},
		},
		{
			args: []string{"dev", "not-context"},
		},
		{
			args: []string{"not-context", "dev"},
		},
		{
			args: []string{"http://k8s.server1.com", "not-a-url"},
		},
	}

	for _, test := range tests {
		err := cmd.RunE(cmd, test.args)
		assert.Error(t, err, "args: %v", test.args)
	}
}

func TestRunWatch(t *testing.T) {
	c := NewMrrCache()
	f := NewTestFactory()
	f.mrrCache = c

	servers := []string{
		"http://k8s.server1.com",
		"http://k8s.server2.com",
	}

	for _, s := range servers {
		kc := NewTestKubeClient()
		kc.baseURL, _ = url.Parse(s)
		f.kubeClients[s] = kc
	}

	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.Flags().Set("interval", "3ms")
	go cmd.RunE(cmd, servers)
	time.Sleep(50 * time.Millisecond)

	for s, kc := range f.kubeClients {
		for _, kind := range []string{"pod"} {
			if kc.watchObjectHits[kind] != 1 {
				t.Errorf("Unexpected number of WatchObject requests for [%s] server [%s]: %v", kind, s, kc.watchObjectHits)
			}
		}
		for _, kind := range []string{"configmap", "namespace", "service", "deployment", "node"} {
			if kc.getObjectHits[kind] < 3 {
				t.Errorf("Expected at least 3 GetObject requests for [%s] server [%s], hits were %v", kind, s, kc.getObjectHits)
			}
		}
	}
}

func TestRunWatchFailedPing(t *testing.T) {
	kc := NewTestKubeClient()
	f := NewTestFactory()
	f.kubeClients[kc.Server().URL] = kc

	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.RunE(cmd, []string{kc.Server().URL})

	assert.Equal(t, kc.pings, 1, "must have pinged server")
}

func TestRunWatchContextMode(t *testing.T) {
	f := NewTestFactory()
	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.Flags().Set("kubeconfig", "test_data/kubeconfig_valid")

	go cmd.RunE(cmd, []string{"prod", "dev"})
	time.Sleep(50 * time.Millisecond)

	//copied from kubeconfig_valid file
	expectedURLs := []string{"https://foo.com", "https://bar.com"}
	actualURLs := []string{}
	for _, kc := range f.kubeClients {
		actualURLs = append(actualURLs, kc.baseURL.String())
	}

	assert.Equal(t, expectedURLs, actualURLs)
}

func TestRunWatchWithOnlyFlag(t *testing.T) {
	f := NewTestFactory()
	cmd := NewWatchCommand(f)
	cmd.Flags().Set("port", "0")
	cmd.Flags().Set("interval", "3ms")
	cmd.Flags().Set("only", "pod,namespace")
	go cmd.RunE(cmd, []string{"http://z.org"})
	time.Sleep(50 * time.Millisecond)

	for _, kc := range f.kubeClients {
		for kind, hits := range kc.watchObjectHits {
			if kind == "pod" && hits < 1 {
				t.Errorf("Expected to hit [%s] at least 3 times, but was [%d]", kind, hits)
			}
			if kind != "pod" && hits > 0 {
				t.Errorf("Did not expect to hit [%s]", kind)
			}
		}

		for kind, hits := range kc.getObjectHits {
			if kind == "namespace" && hits < 3 {
				t.Errorf("Expected to hit [%s] at least 3 times, but was [%d]", kind, hits)
			}
			if kind != "namespace" && hits > 0 {
				t.Errorf("Did not expect to hit [%s]", kind)
			}
		}
	}
}

func TestLoopWatchObjectsFailure(t *testing.T) {
	c := NewMrrCache()
	kind := "o"
	kc := NewTestKubeClient()
	kc.watchObjectError = errors.New("Test Error")
	kc.objectEventsF = func() []*ObjectEvent {
		return []*ObjectEvent{
			&ObjectEvent{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "object"}, TypeMeta: TypeMeta{kind}}},
		}
	}

	loopWatchObjects(c, kc, kind)

	time.Sleep(50 * time.Millisecond)
	if kc.watchObjectHits[kind] < 2 {
		t.Errorf("Not enough WatchObjects calls")
	}

	x := len(c.objects[kc.Server()])
	if x > 1 {
		t.Errorf("Cache must contain only one object, but contains %d", x)
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
		{Added, &KubeObject{TypeMeta: TypeMeta{"other"}, ObjectMeta: ObjectMeta{Name: "pod0"}}},
	}

	loopWatchObjects(c, kc, "does not matter")
	time.Sleep(50 * time.Millisecond)

	//order matters in slice
	expected := []KubeObject{*kc.objectEvents[4].Object, *kc.objectEvents[3].Object, *kc.objectEvents[7].Object}
	actual := c.objects[kc.Server()]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Cache version %+v is not equal to expected %+v", actual, expected)
	}
}

func TestLoopGetObjects(t *testing.T) {
	c := NewMrrCache()
	kc := NewTestKubeClient()
	kind := ""

	finalObjects := []KubeObject{
		{ObjectMeta: ObjectMeta{Name: "a1"}},
		{ObjectMeta: ObjectMeta{Name: "a2"}},
	}
	kc.objectsF = func() []KubeObject {
		if kc.getObjectHits[kind] > 2 {
			return finalObjects
		} else {
			return []KubeObject{
				{ObjectMeta: ObjectMeta{Name: fmt.Sprintf("rand-%d", rand.Intn(999))}},
			}
		}
	}

	loopGetObjects(c, kc, kind, 3*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	actual := c.objects[kc.Server()]
	expected := finalObjects
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected \n%+v \n Got \n%+v", expected, actual)
	}
}
