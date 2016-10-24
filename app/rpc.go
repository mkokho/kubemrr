package app

import (
	"net/rpc"
	"sync"
)

type MrrCache struct {
	objects     map[KubeServer][]KubeObject
	pods        map[string]*Pod
	services    map[string]*Service
	deployments map[string]*Deployment
	mu          *sync.RWMutex
}

func NewMrrCache() *MrrCache {
	c := &MrrCache{}
	c.mu = &sync.RWMutex{}
	c.objects = make(map[KubeServer][]KubeObject)
	c.pods = map[string]*Pod{}
	c.services = map[string]*Service{}
	c.deployments = map[string]*Deployment{}
	return c
}

func (c *MrrCache) Pods(f *MrrFilter, pods *[]Pod) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, p := range c.pods {
		*pods = append(*pods, *p)
	}
	return nil
}

func (c *MrrCache) Services(f *MrrFilter, services *[]Service) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.services {
		*services = append(*services, *s)
	}
	return nil
}

func (c *MrrCache) Deployments(f *MrrFilter, deployments *[]Deployment) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, d := range c.deployments {
		*deployments = append(*deployments, *d)
	}
	return nil
}

func (c *MrrCache) setKubeObjects(server KubeServer, xs []KubeObject) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.objects[server] = xs
}

func (c *MrrCache) updateKubeObject(server KubeServer, o KubeObject) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.objects[server]
	if !ok {
		os = make([]KubeObject, 0)
	}

	found := false
	for i := range os {
		if os[i].Name == o.Name {
			os[i] = o
			found = true
			break
		}
	}

	if !found {
		os = append(os, o)
	}
	c.objects[server] = os
}

func (c *MrrCache) deleteKubeObject(server KubeServer, o KubeObject) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.objects[server]
	if !ok {
		return
	}

	idx := -1
	for i := range os {
		if os[i].Name == o.Name {
			idx = i
			break
		}
	}

	if idx >= 0 {
		os = append(os[:idx], os[idx+1:]...)
		c.objects[server] = os
	}
}

func (c *MrrCache) setPods(pods []Pod) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pods = map[string]*Pod{}
	for _, p := range pods {
		c.pods[p.Name] = &p
	}
}

func (c *MrrCache) updatePod(pod *Pod) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pods[pod.ObjectMeta.Name] = pod
}

func (c *MrrCache) removePod(pod *Pod) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.pods, pod.ObjectMeta.Name)
}

func (c *MrrCache) updateService(s *Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[s.ObjectMeta.Name] = s
}

func (c *MrrCache) removeService(s *Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, s.ObjectMeta.Name)
}

func (c *MrrCache) updateDeployment(d *Deployment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deployments[d.ObjectMeta.Name] = d
}

func (c *MrrCache) removeDeployment(d *Deployment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.deployments, d.ObjectMeta.Name)
}

func (c *MrrCache) setServices(services []Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = map[string]*Service{}
	for _, s := range services {
		c.services[s.Name] = &s
	}
}

func (c *MrrCache) setDeployments(deployments []Deployment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deployments = map[string]*Deployment{}
	for _, d := range deployments {
		c.deployments[d.Name] = &d
	}
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
	err := mc.conn.Call("MrrCache.Pods", &MrrFilter{}, &pods)
	return pods, err
}

func (mc *MrrClientDefault) Services() ([]Service, error) {
	var services []Service
	err := mc.conn.Call("MrrCache.Services", &MrrFilter{}, &services)
	return services, err
}

func (mc *MrrClientDefault) Deployments() ([]Deployment, error) {
	var deployments []Deployment
	err := mc.conn.Call("MrrCache.Deployments", &MrrFilter{}, &deployments)
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
