package main

import (
	"bytes"
	"github.com/mkokho/kubemrr/app"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	log "github.com/Sirupsen/logrus"
)

var (
	k8sServer  *httptest.Server
	k8sAddress string
	mux        *http.ServeMux
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	m.Run()
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
	stream(w, []string{`{"type": "ADDED", "object": {"kind":"service", "metadata": {"name": "service1"}}}`})
}
func k8sDeployments(w http.ResponseWriter, r *http.Request) {
	stream(w, []string{`{"type": "ADDED", "object": {"kind":"deployment", "metadata": {"name": "deployment1"}}}`})
}

func startKubernetesServer() {
	mux = http.NewServeMux()
	k8sServer = httptest.NewServer(mux)
	k8sAddress = k8sServer.URL

	mux.HandleFunc("/api/v1/pods", k8sPods)
	mux.HandleFunc("/api/v1/services", k8sServices)
	mux.HandleFunc("/apis/extensions/v1beta1/deployments", k8sDeployments)
}

func stopKubernetesServer() {
	k8sServer.Close()
}

func TestCommands(t *testing.T) {
	startKubernetesServer()
	defer stopKubernetesServer()

	buf := bytes.NewBuffer([]byte{})
	f := app.NewFactory(buf)
	getCmd := app.NewGetCommand(f)
	watchCmd := app.NewWatchCommand(f)
	go watchCmd.Run(watchCmd, []string{k8sAddress})

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
	}

	for _, test := range tests {
		buf.Reset()
		getCmd.Run(getCmd, []string{test.arg})
		if buf.String() != test.output {
			t.Errorf("Getting [%v]: expected [%v], but received [%v]", test.arg, test.output, buf)
		}
	}
}
