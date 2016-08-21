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
	cache.setPods([]Pod{Pod{ObjectMeta: ObjectMeta{Name: "pod1"}}})
	cache.setServices([]Service{Service{ObjectMeta: ObjectMeta{Name: "service1"}}})
	cache.setDeployments([]Deployment{Deployment{ObjectMeta: ObjectMeta{Name: "deployment1"}}})
	rpc.Register(cache)
	rpc.HandleHTTP()
	go http.Serve(l, nil)

	mrrClient, err = f.MrrClient(l.Addr().String())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func TestClientPods(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Pods()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result, cache.pods) {
		t.Errorf("Expected pods %v, found %v", cache.pods, result)
	}
}

func TestClientServices(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Services()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result, cache.services) {
		t.Errorf("Expected services %v, found %v", cache.services, result)
	}
}

func TestClientDeployments(t *testing.T) {
	once.Do(setupRPC)

	result, err := mrrClient.Deployments()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(result, cache.deployments) {
		t.Errorf("Expected deployments %v, found %v", cache.deployments, result)
	}
}
