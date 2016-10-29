package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type KubeClient interface {
	Server() KubeServer
	WatchObjects(kind string, out chan *ObjectEvent) error
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

type TestKubeClient struct {
	baseURL *url.URL

	objectEvents []*ObjectEvent

	watchObjectHits  map[string]int
	watchObjectLock  *sync.RWMutex
	watchObjectError error
}

func NewTestKubeClient() *TestKubeClient {
	kc := &TestKubeClient{}
	kc.baseURL, _ = url.Parse(fmt.Sprintf("random-url-%d", rand.Intn(999)))
	kc.watchObjectLock = &sync.RWMutex{}
	kc.watchObjectHits = map[string]int{}
	return kc
}

func (kc *TestKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
}

func (kc *TestKubeClient) WatchObjects(kind string, out chan *ObjectEvent) error {
	kc.watchObjectLock.Lock()
	kc.watchObjectHits[kind] += 1
	kc.watchObjectLock.Unlock()

	if kc.watchObjectHits[kind] < 5 && kc.watchObjectError != nil {
		return kc.watchObjectError
	}

	for i := range kc.objectEvents {
		out <- kc.objectEvents[i]
	}
	select {}
}
