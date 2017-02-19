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
	Ping() error
	WatchObjects(kind string, out chan *ObjectEvent) error
	GetObjects(kind string) ([]KubeObject, error)
}

type DefaultKubeClient struct {
	client  *http.Client
	baseURL *url.URL
}

//NewKubeClient returns a client that talks to Kubenetes API server.
//It talks to only one server, and uses configuration of the current context in the
//given config
func NewKubeClient(config *Config) KubeClient {
	tlsConfig, _ := config.GenerateTLSConfig()
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: tr}

	url, _ := url.Parse(config.getCurrentCluster().Server)
	return &DefaultKubeClient{
		client:  httpClient,
		baseURL: url,
	}
}

func (kc *DefaultKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
}

func (kc *DefaultKubeClient) Ping() error {
	req, err := kc.newRequest("GET", "/", nil)
	if err != nil {
		return err
	}
	return kc.do(req, nil)
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
	case "service":
		return kc.get("api/v1/services", kind)
	case "deployment":
		return kc.get("/apis/extensions/v1beta1/deployments", kind)
	case "namespace":
		return kc.get("api/v1/namespaces", kind)
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

	if resp.StatusCode >= 300 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte(err.Error())
		}
		return fmt.Errorf("unexpected status for %s %s: %s %s", req.Method, req.URL, resp.Status, string(body))
	}

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
	pings   int

	objectEvents  []*ObjectEvent
	objectEventsF func() []*ObjectEvent

	watchObjectHits  map[string]int
	watchObjectLock  *sync.RWMutex
	watchObjectError error

	objects       []KubeObject
	objectsF      func() []KubeObject
	getObjectHits map[string]int
}

func NewTestKubeClient() *TestKubeClient {
	kc := &TestKubeClient{}
	kc.baseURL, _ = url.Parse(fmt.Sprintf("http://random-url-%d.com", rand.Intn(999)))
	kc.watchObjectLock = &sync.RWMutex{}
	kc.watchObjectHits = map[string]int{}
	kc.objectEventsF = func() []*ObjectEvent { return []*ObjectEvent{} }
	kc.objects = []KubeObject{}
	kc.objectsF = func() []KubeObject { return []KubeObject{} }
	kc.getObjectHits = map[string]int{}
	return kc
}

func (kc *TestKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
}

func (kc *TestKubeClient) Ping() error {
	kc.pings += 1
	return nil
}

func (kc *TestKubeClient) WatchObjects(kind string, out chan *ObjectEvent) error {
	kc.watchObjectLock.Lock()
	kc.watchObjectHits[kind] += 1
	kc.watchObjectLock.Unlock()

	for i := range kc.objectEvents {
		out <- kc.objectEvents[i]
	}

	for _, o := range kc.objectEventsF() {
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

	if len(kc.objects) == 0 {
		return kc.objectsF(), nil
	} else {
		return kc.objects, nil
	}
}
