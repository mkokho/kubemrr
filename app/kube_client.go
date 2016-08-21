package app

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type KubeClient interface {
	BaseURL() *url.URL
	GetPods() ([]Pod, error)
	GetServices() ([]Service, error)
	GetDeployments() ([]Deployment, error)
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
}

func (kc *TestKubeClient) BaseURL() *url.URL {
	return kc.baseURL
}

func (kc *TestKubeClient) GetPods() ([]Pod, error) {
	kc.hitsGetPods += 1
	return []Pod{Pod{ObjectMeta: ObjectMeta{Name: "pod1"}}}, nil
}

func (kc *TestKubeClient) GetServices() ([]Service, error) {
	kc.hitsGetServices += 1
	return []Service{Service{ObjectMeta: ObjectMeta{Name: "service1"}}}, nil
}

func (kc *TestKubeClient) GetDeployments() ([]Deployment, error) {
	kc.hitsGetDeployments += 1
	return []Deployment{Deployment{ObjectMeta: ObjectMeta{Name: "deployment1"}}}, nil
}
