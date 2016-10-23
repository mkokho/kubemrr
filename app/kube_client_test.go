package app

import (
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
	mux.HandleFunc("/apis/extensions/v1beta1/deployments", func(w http.ResponseWriter, r *http.Request) {
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
