package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

type ObjectEvent struct {
	Type   EventType   `json:"type"`
	Object *KubeObject `json:"object"`
}

type ObjectList struct {
	Objects []KubeObject `json:"items"`
}

type KubeClient interface {
	Server() KubeServer
	WatchObjects(kind string, out chan *ObjectEvent) error
	GetObjects(kind string) ([]KubeObject, error)
}

type DefaultKubeClient struct {
	client  *http.Client
	baseURL *url.URL
}

func NewKubeClient(url *url.URL) KubeClient {
	c := &DefaultKubeClient{
		client:  http.DefaultClient,
		baseURL: url,
	}
	return c
}

func (kc *DefaultKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
}

func (kc *DefaultKubeClient) WatchObjects(kind string, out chan *ObjectEvent) error {
	switch kind {
	case "pod":
		return kc.watch("api/v1/pods?watch=true", out)
	case "service":
		return kc.watch("api/v1/services?watch=true", out)
	case "deployment":
		return kc.watch("/apis/extensions/v1beta1/deployments?watch=true", out)
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (kc *DefaultKubeClient) GetObjects(kind string) ([]KubeObject, error) {
	switch kind {
	case "configmap":
		return kc.get("api/v1/configmaps", kind)
	default:
		return []KubeObject{}, fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (kc *DefaultKubeClient) get(url string, kind string) ([]KubeObject, error) {
	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return []KubeObject{}, err
	}

	var list ObjectList
	err = kc.do(req, &list)
	if err != nil {
		return []KubeObject{}, err
	}

	for i := range list.Objects {
		list.Objects[i].Kind = kind
	}

	return list.Objects, nil
}

func (kc *DefaultKubeClient) watch(url string, out chan *ObjectEvent) error {
	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return err
	}

	res, err := kc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to watch pods: %d", res.StatusCode)
	}

	d := json.NewDecoder(res.Body)

	for {
		var event ObjectEvent
		err := d.Decode(&event)

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Could not decode data into pod event: %s", err)
		}

		out <- &event
	}

	return nil
}

func (kc *DefaultKubeClient) newRequest(method string, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	u := kc.baseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *DefaultKubeClient) do(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err == io.EOF {
			err = nil // ignore EOF errors caused by empty response body
		}
	}

	return err
}

type TestKubeClient struct {
	baseURL *url.URL

	objectEvents     []*ObjectEvent
	makeObjectEvents func() []*ObjectEvent

	watchObjectHits  map[string]int
	watchObjectLock  *sync.RWMutex
	watchObjectError error

	objects       []KubeObject
	getObjectHits map[string]int
}

func NewTestKubeClient() *TestKubeClient {
	kc := &TestKubeClient{}
	kc.baseURL, _ = url.Parse(fmt.Sprintf("random-url-%d", rand.Intn(999)))
	kc.watchObjectLock = &sync.RWMutex{}
	kc.watchObjectHits = map[string]int{}
	kc.makeObjectEvents = func() []*ObjectEvent { return []*ObjectEvent{} }
	kc.objects = []KubeObject{}
	kc.getObjectHits = map[string]int{}
	return kc
}

func (kc *TestKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
}

func (kc *TestKubeClient) WatchObjects(kind string, out chan *ObjectEvent) error {
	kc.watchObjectLock.Lock()
	kc.watchObjectHits[kind] += 1
	kc.watchObjectLock.Unlock()

	for i := range kc.objectEvents {
		out <- kc.objectEvents[i]
	}

	for _, o := range kc.makeObjectEvents() {
		out <- o
	}

	if kc.watchObjectHits[kind] < 5 && kc.watchObjectError != nil {
		return kc.watchObjectError
	}

	select {}
}

func (kc *TestKubeClient) GetObjects(kind string) ([]KubeObject, error) {
	kc.watchObjectLock.Lock()
	kc.getObjectHits[kind] += 1
	kc.watchObjectLock.Unlock()

	return kc.objects, nil
}
