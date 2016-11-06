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
					if c.objects[ks] == nil {
						c.objects[ks] = make([]KubeObject, 0)
					}

					o := KubeObject{TypeMeta{kind}, ObjectMeta{Name: s + "-" + name, Namespace: ns}}
					c.objects[ks] = append(c.objects[ks], o)
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
			filter: MrrFilter{"server_other", "ns1", "pod"},
		},
		{
			filter: MrrFilter{"server1", "ns_other", "pod"},
		},
		{
			filter: MrrFilter{"server1", "ns1", "pod_other"},
		},
		{
			filter: MrrFilter{"SERVER1", "ns1", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{"server2:8443", "NS1", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server2-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{"server1", "ns2", "POD"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns2", ""}},
			},
		},
		{
			filter: MrrFilter{"server1", "ns1", "service"},
			expected: []KubeObject{
				{TypeMeta{"service"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"service"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"service"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{"server1", "ns1", "deployment"},
			expected: []KubeObject{
				{TypeMeta{"deployment"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"deployment"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"deployment"}, ObjectMeta{"server1-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{"", "ns1", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server2-c", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server3-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server3-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server3-c", "ns1", ""}},
			},
		},
		{
			filter: MrrFilter{"server1", "", "pod"},
			expected: []KubeObject{
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns1", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns2", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-a", "ns3", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-b", "ns3", ""}},
				{TypeMeta{"pod"}, ObjectMeta{"server1-c", "ns3", ""}},
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

func TestDeleteKubeObjects(t *testing.T) {
	c := NewMrrCache()
	s := KubeServer{"s"}
	o1 := KubeObject{TypeMeta: TypeMeta{"x"}, ObjectMeta: ObjectMeta{Name: "x1"}}
	o2 := KubeObject{TypeMeta: TypeMeta{"y"}, ObjectMeta: ObjectMeta{Name: "y1"}}
	c.updateKubeObject(s, o1)
	c.updateKubeObject(s, o2)

	c.deleteKubeObjects(s, "y")
	if !reflect.DeepEqual(c.objects[s], []KubeObject{o1}) {
		t.Errorf("Cache should contain only %+v, but it contains %+v", o1, c.objects[s])
	}
}

func TestUpdateKubeObject(t *testing.T) {
	c := NewMrrCache()
	s := KubeServer{"s"}

	expected := []KubeObject{
		{TypeMeta: TypeMeta{"x"}, ObjectMeta: ObjectMeta{Name: "x1"}},
		{TypeMeta: TypeMeta{"y"}, ObjectMeta: ObjectMeta{Name: "x1"}},
		{TypeMeta: TypeMeta{"y"}, ObjectMeta: ObjectMeta{Name: "x1", Namespace: "ns2"}},
	}

	for i := range expected {
		c.updateKubeObject(s, expected[i])
	}

	if !reflect.DeepEqual(c.objects[s], expected) {
		t.Errorf("Cache should all %d obejcts, but it contains %+v", len(expected), c.objects[s])
	}
}
