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
	client *KubeClient
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	url, _ := url.Parse(server.URL)
	client = NewKubeClient()
	client.BaseURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func TestFetchPods(t *testing.T) {
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

	pods, err := client.getPods()
	if err != nil {
		t.Errorf("getPods returned error: %v", err)
	}

	want := []Pod{
		Pod{ObjectMeta: ObjectMeta{Name: "pod1"}},
		Pod{ObjectMeta: ObjectMeta{Name: "pod2"}},
	}

	if !reflect.DeepEqual(pods, want) {
		t.Errorf("getPods returned %+v, want %+v", pods, want)
	}
}

func TestFetchServices(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			{
				"items": [
					{ "metadata": { "name": "svc1" } },
					{ "metadata": { "name": "svc2" } }
				]
			}`)
	},
	)

	svc, err := client.getServices()
	if err != nil {
		t.Errorf("getServices returned error: %v", err)
	}

	want := []Service{
		Service{ObjectMeta: ObjectMeta{Name: "svc1"}},
		Service{ObjectMeta: ObjectMeta{Name: "svc2"}},
	}

	if !reflect.DeepEqual(svc, want) {
		t.Errorf("getPods returned %+v, want %+v", svc, want)
	}
}
