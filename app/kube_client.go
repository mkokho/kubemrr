package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
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
	Server() KubeServer
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

func (kc *DefaultKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
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
	req, err := kc.newRequest("GET", "/apis/extensions/v1beta1/deployments?watch=true", nil)
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
	baseURL            *url.URL
	hitsGetPods        int
	hitsGetServices    int
	hitsGetDeployments int

	podEvents        []*PodEvent
	serviceEvents    []*ServiceEvent
	deploymentEvents []*DeploymentEvent

	hits   map[string]int
	errors map[string]error
}

func NewTestKubeClient() *TestKubeClient {
	kc := &TestKubeClient{}
	kc.baseURL, _ = url.Parse(fmt.Sprintf("random-url-%d", rand.Intn(999)))
	kc.hits = map[string]int{}
	kc.errors = map[string]error{}
	return kc
}

func (kc *TestKubeClient) Server() KubeServer {
	return KubeServer{kc.baseURL.String()}
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
	select {}
}

func (kc *TestKubeClient) WatchServices(out chan *ServiceEvent) error {
	kc.hits["WatchServices"] += 1
	if kc.hits["WatchServices"] < 5 && kc.errors["WatchServices"] != nil {
		return kc.errors["WatchServices"]
	}

	for i := range kc.serviceEvents {
		out <- kc.serviceEvents[i]
	}
	select {}
}

func (kc *TestKubeClient) WatchDeployments(out chan *DeploymentEvent) error {
	kc.hits["WatchDeployments"] += 1
	if kc.hits["WatchDeployments"] < 5 && kc.errors["WatchDeployments"] != nil {
		return kc.errors["WatchDeployments"]
	}

	for i := range kc.deploymentEvents {
		out <- kc.deploymentEvents[i]
	}
	select {}
}
