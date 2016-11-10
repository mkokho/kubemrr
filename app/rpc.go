package app

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/rpc"
	"sort"
	"strings"
	"sync"
)

type MrrFilter struct {
	Server    string
	Namespace string
	Kind      string
}

type MrrCache struct {
	objects map[KubeServer][]KubeObject
	mu      *sync.RWMutex
}

func NewMrrCache() *MrrCache {
	c := &MrrCache{}
	c.mu = &sync.RWMutex{}
	c.objects = make(map[KubeServer][]KubeObject)
	return c
}

func (c *MrrCache) Objects(f *MrrFilter, os *[]KubeObject) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	log.WithField("filter", f).Debug("Received request for objects")

	if f == nil {
		return errors.New("Cannot find pods with nil filter")
	}

	keys := KubeServers{}
	for k, _ := range c.objects {
		if f.Server == "" || strings.EqualFold(trimPort(f.Server), trimPort(k.URL)) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		log.WithField("server", f.Server).Error("unknown server")
		return fmt.Errorf("Unknown server %s", f.Server)
	}

	res := []KubeObject{}
	sort.Sort(keys)
	for _, k := range keys {
		for _, o := range c.objects[k] {
			if strings.EqualFold(o.Kind, f.Kind) &&
				(f.Namespace == "" || strings.EqualFold(o.Namespace, f.Namespace)) {
				res = append(res, o)
			}
		}
	}
	log.WithField("filter", f).WithField("objects", res).Debug("Returning result for objects")
	*os = res
	return nil
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
		if os[i].Name == o.Name && os[i].Namespace == o.Namespace && os[i].Kind == o.Kind {
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
		if os[i].Name == o.Name && os[i].Namespace == o.Namespace {
			idx = i
			break
		}
	}

	if idx >= 0 {
		os = append(os[:idx], os[idx+1:]...)
		c.objects[server] = os
	}
}

func (c *MrrCache) deleteKubeObjects(s KubeServer, kind string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.objects[s]
	if !ok {
		return
	}

	newObjects := []KubeObject{}
	for i := range os {
		if os[i].Kind != kind {
			newObjects = append(newObjects, os[i])
		}
	}

	c.objects[s] = newObjects
}

func trimPort(url string) string {
	i := strings.LastIndex(url, ":")
	if i < 7 {
		return url
	} else {
		return url[:i]
	}
}

type MrrClient interface {
	Objects(f MrrFilter) ([]KubeObject, error)
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

func (mc *MrrClientDefault) Objects(f MrrFilter) ([]KubeObject, error) {
	var os []KubeObject
	err := mc.conn.Call("MrrCache.Objects", f, &os)
	return os, err
}

type TestMirrorClient struct {
	err        error
	lastFilter MrrFilter
	objects    []KubeObject
}

func (mc *TestMirrorClient) Objects(f MrrFilter) ([]KubeObject, error) {
	mc.lastFilter = f
	return mc.objects, mc.err
}
