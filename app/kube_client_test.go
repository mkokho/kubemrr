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
		&ObjectEvent{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "first"}}},
		&ObjectEvent{Modified, &KubeObject{ObjectMeta: ObjectMeta{Name: "second"}}},
		&ObjectEvent{Deleted, &KubeObject{ObjectMeta: ObjectMeta{Name: "last"}}},
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

	inEvents := make(chan *ObjectEvent, 10)
	err := client.WatchObjects("pod", inEvents)
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
		&ObjectEvent{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "first"}}},
		&ObjectEvent{Modified, &KubeObject{ObjectMeta: ObjectMeta{Name: "second"}}},
		&ObjectEvent{Deleted, &KubeObject{ObjectMeta: ObjectMeta{Name: "last"}}},
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

	inEvents := make(chan *ObjectEvent, 10)
	err := client.WatchObjects("service", inEvents)
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
		&ObjectEvent{Added, &KubeObject{ObjectMeta: ObjectMeta{Name: "first"}}},
		&ObjectEvent{Modified, &KubeObject{ObjectMeta: ObjectMeta{Name: "second"}}},
		&ObjectEvent{Deleted, &KubeObject{ObjectMeta: ObjectMeta{Name: "last"}}},
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

	inEvents := make(chan *ObjectEvent, 10)
	err := client.WatchObjects("deployment", inEvents)
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

func TestGetConfigmaps(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/configmaps", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "x1" } },
					{ "metadata": { "name": "x2" } }
				]
			}`)
	},
	)

	res, err := client.GetObjects("configmap")
	if err != nil {
		t.Errorf("GetServices returned error: %v", err)
	}

	expected := []KubeObject{
		{TypeMeta: TypeMeta{"configmap"}, ObjectMeta: ObjectMeta{Name: "x1"}},
		{TypeMeta: TypeMeta{"configmap"}, ObjectMeta: ObjectMeta{Name: "x2"}},
	}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("Expected %+v, got %+v", expected, res)
	}
}

func TestGetNamespaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "x1" } },
					{ "metadata": { "name": "x2" } }
				]
			}`)
	},
	)

	res, err := client.GetObjects("namespace")
	if err != nil {
		t.Errorf("GetServices returned error: %v", err)
	}

	expected := []KubeObject{
		{TypeMeta: TypeMeta{"namespace"}, ObjectMeta: ObjectMeta{Name: "x1"}},
		{TypeMeta: TypeMeta{"namespace"}, ObjectMeta: ObjectMeta{Name: "x2"}},
	}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("Expected %+v, got %+v", expected, res)
	}
}
