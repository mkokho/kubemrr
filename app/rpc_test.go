package app

import (
	"log"
	"net"
	"reflect"
	"sync"
	"testing"
)

var (
	cache     *MrrCache
	mrrClient *MrrClientDefault
	once      sync.Once
)

func setupRPC() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}

	cache = NewMrrCache()
	cache.setPods([]Pod{Pod{ObjectMeta: ObjectMeta{Name: "pod1"}}})
	go ServeMrrCache(l, cache)

	mrrClient, err = NewMrrClient(l.Addr().String())
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
