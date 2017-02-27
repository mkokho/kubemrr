package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mkokho/kubemrr/app"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	k8sServer  *httptest.Server
	k8sAddress string
	mux        *http.ServeMux
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	code := m.Run()
	os.Exit(code)
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
	time.Sleep(50 * time.Millisecond)
}

func k8sPods(w http.ResponseWriter, r *http.Request) {
	stream(w, []string{`{"type": "ADDED", "object": {"kind":"pod", "metadata": {"name": "pod1"}}}`})
}

func k8sServices(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{ "items": [ { "metadata": { "name": "service1" } } ] }`)
}
func k8sDeployments(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{ "items": [ { "metadata": { "name": "deployment1" } } ] }`)
}

func k8sConfigmaps(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, ` { "items": [ { "metadata": { "name": "configmap1" } } ] }`)
}

func k8sNamespaces(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, ` { "items": [ { "metadata": { "name": "namespace1" } } ] }`)
}
func k8sNodes(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, ` { "items": [ { "metadata": { "name": "node1" } } ] }`)
}

func ok(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `OK`)
}

func startKubernetesServer() {
	mux = http.NewServeMux()
	k8sServer = httptest.NewServer(mux)
	k8sAddress = k8sServer.URL

	mux.HandleFunc("/", ok)
	mux.HandleFunc("/api/v1/pods", k8sPods)
	mux.HandleFunc("/api/v1/services", k8sServices)
	mux.HandleFunc("/api/v1/configmaps", k8sConfigmaps)
	mux.HandleFunc("/api/v1/namespaces", k8sNamespaces)
	mux.HandleFunc("/api/v1/nodes", k8sNodes)
	mux.HandleFunc("/apis/extensions/v1beta1/deployments", k8sDeployments)
}

func stopKubernetesServer() {
	k8sServer.Close()
}

func TestCommands(t *testing.T) {
	startKubernetesServer()
	defer stopKubernetesServer()

	buf := bytes.NewBuffer([]byte{})
	f := app.NewFactory(buf, &app.Config{})
	getCmd := app.NewGetCommand(f)
	getCmd.Flags().Set("port", "39000")
	watchCmd := app.NewWatchCommand(f)
	watchCmd.Flags().Set("port", "39000")
	go watchCmd.RunE(watchCmd, []string{k8sAddress})

	time.Sleep(1 * time.Millisecond)

	tests := []struct {
		arg    string
		output string
	}{
		{
			arg:    "pod",
			output: "pod1",
		},
		{
			arg:    "service",
			output: "service1",
		},
		{
			arg:    "deployment",
			output: "deployment1",
		},
		{
			arg:    "configmap",
			output: "configmap1",
		},
		{
			arg:    "namespace",
			output: "namespace1",
		},
		{
			arg:    "node",
			output: "node1",
		},
	}

	for _, test := range tests {
		buf.Reset()
		getCmd.RunE(getCmd, []string{test.arg})
		if buf.String() != test.output {
			t.Errorf("Getting [%v]: expected [%v], but received [%v]", test.arg, test.output, buf)
		}
	}
}
