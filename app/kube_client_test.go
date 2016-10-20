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
		{ObjectMeta: ObjectMeta{Name: "pod1"}},
		{ObjectMeta: ObjectMeta{Name: "pod2"}},
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
		{ObjectMeta: ObjectMeta{Name: "service1"}},
		{ObjectMeta: ObjectMeta{Name: "service2"}},
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
		{ObjectMeta: ObjectMeta{Name: "deployment1"}},
		{ObjectMeta: ObjectMeta{Name: "deployment2"}},
	}

	if !reflect.DeepEqual(svc, want) {
		t.Errorf("GetDeployments returned %+v, want %+v", svc, want)
	}
}

func stream(w http.ResponseWriter, items []string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("need flusher!")
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for _, item := range items {
		_, err := w.Write([]byte(item))
		if err != nil {
			panic(err)
		}
		flusher.Flush()
	}
}

func TestWatchPods(t *testing.T) {
	events := []interface{}{
		&PodEvent{Added, &Pod{ObjectMeta: ObjectMeta{Name: "first"}}},
		&PodEvent{Modified, &Pod{ObjectMeta: ObjectMeta{Name: "second"}}},
		&PodEvent{Deleted, &Pod{ObjectMeta: ObjectMeta{Name: "last"}}},
	}

	setup()
	defer teardown()
	mux.HandleFunc("/api/v1/pods", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("watch") != "true" {
			t.Errorf("URL must have parameter `?watch=true`")
		}
		stream(w, []string{
			`{"type": "ADDED", "object": {"metadata": {"name": "first"}}}`,
			`{"type": "MODIFIED", "object": {"metadata": {"name": "second"}}}`,
			`{"type": "DELETED", "object": {"metadata": {"name": "last"}}}`,
		})
	},
	)

	inEvents := make(chan *PodEvent, 10)
	err := client.WatchPods(inEvents)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, expectedEvent := range events {
		actualEvent := <-inEvents

		if !reflect.DeepEqual(expectedEvent, actualEvent) {
			t.Errorf("Expected %v, received %v", expectedEvent, actualEvent)
		}
	}
}

func TestWatchServices(t *testing.T) {
	events := []interface{}{
		&ServiceEvent{Added, &Service{ObjectMeta: ObjectMeta{Name: "first"}}},
		&ServiceEvent{Modified, &Service{ObjectMeta: ObjectMeta{Name: "second"}}},
		&ServiceEvent{Deleted, &Service{ObjectMeta: ObjectMeta{Name: "last"}}},
	}

	setup()
	defer teardown()
	mux.HandleFunc("/api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("watch") != "true" {
			t.Errorf("URL must have parameter `?watch=true`")
		}
		stream(w, []string{
			`{"type": "ADDED", "object": {"metadata": {"name": "first"}}}`,
			`{"type": "MODIFIED", "object": {"metadata": {"name": "second"}}}`,
			`{"type": "DELETED", "object": {"metadata": {"name": "last"}}}`,
		})
	},
	)

	inEvents := make(chan *ServiceEvent, 10)
	err := client.WatchServices(inEvents)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, expectedEvent := range events {
		actualEvent := <-inEvents

		if !reflect.DeepEqual(expectedEvent, actualEvent) {
			t.Errorf("Expected %v, received %v", expectedEvent, actualEvent)
		}
	}
}

func TestWatchDeployments(t *testing.T) {
	events := []interface{}{
		&DeploymentEvent{Added, &Deployment{ObjectMeta: ObjectMeta{Name: "first"}}},
		&DeploymentEvent{Modified, &Deployment{ObjectMeta: ObjectMeta{Name: "second"}}},
		&DeploymentEvent{Deleted, &Deployment{ObjectMeta: ObjectMeta{Name: "last"}}},
	}

	setup()
	defer teardown()
	mux.HandleFunc("/api/v1/deployments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("watch") != "true" {
			t.Errorf("URL must have parameter `?watch=true`")
		}
		stream(w, []string{
			`{"type": "ADDED", "object": {"metadata": {"name": "first"}}}`,
			`{"type": "MODIFIED", "object": {"metadata": {"name": "second"}}}`,
			`{"type": "DELETED", "object": {"metadata": {"name": "last"}}}`,
		})
	},
	)

	inEvents := make(chan *DeploymentEvent, 10)
	err := client.WatchDeployments(inEvents)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, expectedEvent := range events {
		actualEvent := <-inEvents

		if !reflect.DeepEqual(expectedEvent, actualEvent) {
			t.Errorf("Expected %v, received %v", expectedEvent, actualEvent)
		}
	}
}
