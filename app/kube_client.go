package app

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type KubeClient struct {
	client  *http.Client
	BaseURL *url.URL
}

func NewKubeClient() *KubeClient {
	c := &KubeClient{client: http.DefaultClient}
	return c
}

func (kc *KubeClient) getPods() ([]Pod, error) {
	req, err := kc.newRequest("GET", "api/v1/pods", nil)
	if err != nil {
		return nil, err
	}

	podList := new(PodList)
	err = kc.Do(req, podList)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

func (kc *KubeClient) getServices() ([]Service, error) {
	req, err := kc.newRequest("GET", "api/v1/services", nil)
	if err != nil {
		return nil, err
	}

	svcList := new(ServiceList)
	err = kc.Do(req, svcList)
	if err != nil {
		return nil, err
	}

	return svcList.Items, nil
}

func (kc *KubeClient) getDeployments() ([]Deployment, error) {
	req, err := kc.newRequest("GET", "/apis/extensions/v1beta1/deployments", nil)
	if err != nil {
		return nil, err
	}

	ds := new(DeploymentList)
	err = kc.Do(req, ds)
	if err != nil {
		return nil, err
	}

	return ds.Items, nil
}

func (kc *KubeClient) newRequest(method string, urlStr string, body interface{}) (*http.Request, error) {
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

	u := kc.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *KubeClient) Do(req *http.Request, v interface{}) error {
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
