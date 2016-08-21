package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	client KubeClient
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	url, _ := url.Parse(server.URL)
	f := &DefaultFactory{}
	client = f.KubeClient(url)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestGetPods(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/pods", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "pod1" } },
					{ "metadata": { "name": "pod2" } }
				]
			}`)
	},
	)

	pods, err := client.GetPods()
	if err != nil {
		t.Errorf("GetPods returned error: %v", err)
	}

	want := []Pod{
		Pod{ObjectMeta: ObjectMeta{Name: "pod1"}},
		Pod{ObjectMeta: ObjectMeta{Name: "pod2"}},
	}

	if !reflect.DeepEqual(pods, want) {
		t.Errorf("GetPods returned %+v, want %+v", pods, want)
	}
}

func TestGetServices(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "service1" } },
					{ "metadata": { "name": "service2" } }
				]
			}`)
	},
	)

	services, err := client.GetServices()
	if err != nil {
		t.Errorf("GetServices returned error: %v", err)
	}

	want := []Service{
		Service{ObjectMeta: ObjectMeta{Name: "service1"}},
		Service{ObjectMeta: ObjectMeta{Name: "service2"}},
	}

	if !reflect.DeepEqual(services, want) {
		t.Errorf("GetServices returned %+v, want %+v", services, want)
	}
}

func TestGetDeployments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/apis/extensions/v1beta1/deployments", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "deployment1" } },
					{ "metadata": { "name": "deployment2" } }
				]
			}`)
	},
	)

	svc, err := client.GetDeployments()
	if err != nil {
		t.Errorf("GetDeployments returned error: %v", err)
	}

	want := []Deployment{
		Deployment{ObjectMeta: ObjectMeta{Name: "deployment1"}},
		Deployment{ObjectMeta: ObjectMeta{Name: "deployment2"}},
	}

	if !reflect.DeepEqual(svc, want) {
		t.Errorf("GetDeployments returned %+v, want %+v", svc, want)
	}
}
