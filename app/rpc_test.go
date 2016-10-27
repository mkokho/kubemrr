package app

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
	"sync"
	"testing"
)

var (
	cache     *MrrCache
	mrrClient MrrClient
	once      sync.Once
)

func setupRPC() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}
	f := &DefaultFactory{}
	cache = f.MrrCache()
	fillCache(cache)
	cache.setPods([]Pod{{ObjectMeta: ObjectMeta{Name: "pod1"}}})
	cache.setServices([]Service{{ObjectMeta: ObjectMeta{Name: "service1"}}})
	cache.setDeployments([]Deployment{{ObjectMeta: ObjectMeta{Name: "deployment1"}}})
	rpc.Register(cache)
	rpc.HandleHTTP()
	go http.Serve(l, nil)

	mrrClient, err = f.MrrClient(l.Addr().String())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func fillCache(c *MrrCache) {
	for _, s := range []string{"server1", "server2", "server3"} {
		for _, ns := range []string{"ns1", "ns2", "ns3"} {
			for _, kind := range []string{"pod", "service", "deployment"} {
				for _, name := range []string{"a", "b", "c"} {
					ks := KubeServer{s}
					if cache.objects[ks] == nil {
						cache.objects[ks] = make([]KubeObject, 27)
					}

					o := KubeObject{TypeMeta{kind}, ObjectMeta{Name: s + "-" + name, Namespace: ns}}
					cache.objects[ks] = append(cache.objects[ks], o)
				}
			}
		}
	}
}

func TestClientObjects(t *testing.T) {
	once.Do(setupRPC)

	tests := []struct {
		filter   MrrFilter
		expected []KubeObject
	}{
		{
			filter: MrrFilter{},
		},
		{
			filter: MrrFilter{KubeServer{"server_other"}, "ns1", "pod"},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns_other", "pod"},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns1", "pod_other"},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns1", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{KubeServer{"server2"}, "ns1", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server2-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns2", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns2", ""}},
			},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns1", "service"},
			expected: []KubeObject{
				{TypeMeta{"service"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"service"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"service"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{KubeServer{"server1"}, "ns1", "deployment"},
			expected: []KubeObject{
				{TypeMeta{"deployment"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"deployment"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"deployment"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
	}

	for i, test := range tests {
		actual, err := mrrClient.Objects(test.filter)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("Test %d: expected %#v, found %#v", i, test.expected, actual)
		}
	}
}

func TestClientPods(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Pods()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result[0], *cache.pods["pod1"]) {
		t.Errorf("Expected pods %v, found %v", cache.pods, result)
	}
}

func TestClientServices(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Services()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result[0], *cache.services["service1"]) {
		t.Errorf("Expected services %v, found %v", cache.services, result)
	}
}

func TestClientDeployments(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Deployments()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result[0], *cache.deployments["deployment1"]) {
		t.Errorf("Expected deployments %v, found %v", cache.deployments, result)
	}
}
