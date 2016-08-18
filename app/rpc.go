package app

import ()
import (
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

func ServeMrrCache(l net.Listener, cache *MrrCache) error {
	rpc.Register(cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}

type MrrCache struct {
	pods     []Pod
	services []Service
	mu       *sync.RWMutex
}

func NewMrrCache() *MrrCache {
	return &MrrCache{
		mu: &sync.RWMutex{},
	}
}

func (c *MrrCache) Pods(f *Filter, pods *[]Pod) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	*pods = c.pods
	return nil
}

func (c *MrrCache) Services(f *Filter, services *[]Service) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	*services = c.services
	return nil
}

func (c *MrrCache) setPods(pods []Pod) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pods = pods
}

func (c *MrrCache) setServices(services []Service) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = services
}

type MrrClient interface {
	Pods() ([]Pod, error)
}

type MrrClientDefault struct {
	conn *rpc.Client
}

func NewMrrClient(address string) (*MrrClientDefault, error) {
	connection, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}

	return &MrrClientDefault{conn: connection}, nil
}

func (mc *MrrClientDefault) Pods() ([]Pod, error) {
	var pods []Pod
	err := mc.conn.Call("MrrCache.Pods", &Filter{}, &pods)
	return pods, err
}

func (mc *MrrClientDefault) Services() ([]Service, error) {
	var services []Service
	err := mc.conn.Call("MrrCache.Services", &Filter{}, &services)
	return services, err
}
