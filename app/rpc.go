package app

import (
	"net/rpc"
	"sync"
)

type MrrCache struct {
	pods        []Pod
	services    []Service
	deployments []Deployment
	mu          *sync.RWMutex
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

func (c *MrrCache) Deployments(f *Filter, deployments *[]Deployment) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	*deployments = c.deployments
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

func (c *MrrCache) setDeployments(deployments []Deployment) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deployments = deployments
}

type MrrClient interface {
	Pods() ([]Pod, error)
	Services() ([]Service, error)
	Deployments() ([]Deployment, error)
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

func (mc *MrrClientDefault) Deployments() ([]Deployment, error) {
	var deployments []Deployment
	err := mc.conn.Call("MrrCache.Deployments", &Filter{}, &deployments)
	return deployments, err
}

type TestMirrorClient struct {
	err         error
	pods        []Pod
	services    []Service
	deployments []Deployment
}

func (mc *TestMirrorClient) Pods() ([]Pod, error) {
	return mc.pods, mc.err
}

func (mc *TestMirrorClient) Services() ([]Service, error) {
	return mc.services, mc.err
}

func (mc *TestMirrorClient) Deployments() ([]Deployment, error) {
	return mc.deployments, mc.err
}
