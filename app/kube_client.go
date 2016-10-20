package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

type PodEvent struct {
	Type EventType `json:"type"`
	Pod  *Pod      `json:"object"`
}

type ServiceEvent struct {
	Type    EventType `json:"type"`
	Service *Service  `json:"object"`
}

type DeploymentEvent struct {
	Type       EventType   `json:"type"`
	Deployment *Deployment `json:"object"`
}

type KubeClient interface {
	BaseURL() *url.URL
	GetPods() ([]Pod, error)
	GetServices() ([]Service, error)
	GetDeployments() ([]Deployment, error)
	WatchPods(out chan *PodEvent) error
	WatchServices(out chan *ServiceEvent) error
	WatchDeployments(out chan *DeploymentEvent) error
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

func (kc *DefaultKubeClient) BaseURL() *url.URL {
	return kc.baseURL
}

func (kc *DefaultKubeClient) GetPods() ([]Pod, error) {
	req, err := kc.newRequest("GET", "api/v1/pods", nil)
	if err != nil {
		return nil, err
	}

	podList := new(PodList)
	err = kc.do(req, podList)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

func (kc *DefaultKubeClient) WatchPods(out chan *PodEvent) error {
	req, err := kc.newRequest("GET", "api/v1/pods?watch=true", nil)
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
		var event PodEvent
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

func (kc *DefaultKubeClient) WatchServices(out chan *ServiceEvent) error {
	req, err := kc.newRequest("GET", "api/v1/services?watch=true", nil)
	if err != nil {
		return err
	}

	res, err := kc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to watch services: %d", res.StatusCode)
	}

	d := json.NewDecoder(res.Body)

	for {
		var event ServiceEvent
		err := d.Decode(&event)

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Could not decode data into service event: %s", err)
		}

		out <- &event
	}

	return nil
}

func (kc *DefaultKubeClient) WatchDeployments(out chan *DeploymentEvent) error {
	req, err := kc.newRequest("GET", "api/v1/deployments?watch=true", nil)
	if err != nil {
		return err
	}

	res, err := kc.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to watch deployments: %d", res.StatusCode)
	}

	d := json.NewDecoder(res.Body)

	for {
		var event DeploymentEvent
		err := d.Decode(&event)

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Could not decode data into deployment event: %s", err)
		}

		out <- &event
	}

	return nil
}

func (kc *DefaultKubeClient) GetServices() ([]Service, error) {
	req, err := kc.newRequest("GET", "api/v1/services", nil)
	if err != nil {
		return nil, err
	}

	svcList := new(ServiceList)
	err = kc.do(req, svcList)
	if err != nil {
		return nil, err
	}

	return svcList.Items, nil
}

func (kc *DefaultKubeClient) GetDeployments() ([]Deployment, error) {
	req, err := kc.newRequest("GET", "/apis/extensions/v1beta1/deployments", nil)
	if err != nil {
		return nil, err
	}

	ds := new(DeploymentList)
	err = kc.do(req, ds)
	if err != nil {
		return nil, err
	}

	return ds.Items, nil
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
		//io.Copy(os.Stdout, resp.Body)
		err = json.NewDecoder(resp.Body).Decode(v)
		if err == io.EOF {
			err = nil // ignore EOF errors caused by empty response body
		}
	}

	return err
}

type TestKubeClient struct {
	baseURL            *url.URL
	hitsGetPods        int
	hitsGetServices    int
	hitsGetDeployments int

	podEvents     []*PodEvent
	serviceEvents []*ServiceEvent

	hits   map[string]int
	errors map[string]error
}

func NewTestKubeClient() *TestKubeClient {
	kc := &TestKubeClient{}
	kc.hits = map[string]int{}
	kc.errors = map[string]error{}
	return kc
}

func (kc *TestKubeClient) BaseURL() *url.URL {
	return kc.baseURL
}

func (kc *TestKubeClient) GetPods() ([]Pod, error) {
	kc.hitsGetPods += 1
	return []Pod{{ObjectMeta: ObjectMeta{Name: "pod1"}}}, nil
}

func (kc *TestKubeClient) GetServices() ([]Service, error) {
	kc.hitsGetServices += 1
	return []Service{{ObjectMeta: ObjectMeta{Name: "service1"}}}, nil
}

func (kc *TestKubeClient) GetDeployments() ([]Deployment, error) {
	kc.hitsGetDeployments += 1
	return []Deployment{{ObjectMeta: ObjectMeta{Name: "deployment1"}}}, nil
}

func (kc *TestKubeClient) WatchPods(out chan *PodEvent) error {
	kc.hits["WatchPods"] += 1
	if kc.hits["WatchPods"] < 5 && kc.errors["WatchPods"] != nil {
		return kc.errors["WatchPods"]
	}

	for i := range kc.podEvents {
		out <- kc.podEvents[i]
	}
	for {
	}
}

func (kc *TestKubeClient) WatchServices(out chan *ServiceEvent) error {
	kc.hits["WatchServices"] += 1
	if kc.hits["WatchServices"] < 5 && kc.errors["WatchServices"] != nil {
		return kc.errors["WatchServices"]
	}

	for i := range kc.serviceEvents {
		out <- kc.serviceEvents[i]
	}
	for {
	}
}

func (kc *TestKubeClient) WatchDeployments(out chan *DeploymentEvent) error {
	return nil
}
