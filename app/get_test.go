package app

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRunGetInvalidArgs(t *testing.T) {
	tests := []struct {
		args   []string
		output string
	}{
		{
			args:   []string{},
			output: "no resource",
		},
		{
			args:   []string{"1", "2"},
			output: "one argument",
		},
		{
			args:   []string{"k8s-resource"},
			output: "unsupported resource type",
		},
	}

	f := &TestFactory{}
	cmd := NewGetCommand(f)

	for i, test := range tests {
		err := cmd.RunE(cmd, test.args)
		if err == nil {
			t.Errorf("Test %d: expected: %v, no error was returned", i, test.output)
		}

		if !strings.Contains(err.Error(), test.output) {
			t.Errorf("Test %d: output [%v] does not contains expected [%v]", i, err.Error(), test.output)
		}
	}
}

func TestRunGet(t *testing.T) {
	tc := &TestMirrorClient{
		objects: []KubeObject{
			{ObjectMeta: ObjectMeta{Name: "o1"}},
			{ObjectMeta: ObjectMeta{Name: "o2"}},
		},
	}
	buf := bytes.NewBuffer([]byte{})
	f := &TestFactory{mrrClient: tc, stdOut: buf}
	cmd := NewGetCommand(f)

	expectedOutput := "o1 o2"
	tests := []struct {
		aliases        []string
		expectedFilter MrrFilter
	}{
		{
			aliases:        []string{"po", "pod", "pods"},
			expectedFilter: MrrFilter{Kind: "pod"},
		},
		{
			aliases:        []string{"svc", "service", "services"},
			expectedFilter: MrrFilter{Kind: "service"},
		},
		{
			aliases:        []string{"deployment", "deployments"},
			expectedFilter: MrrFilter{Kind: "deployment"},
		},
		{
			aliases:        []string{"configmap", "configmaps"},
			expectedFilter: MrrFilter{Kind: "configmap"},
		},
		{
			aliases:        []string{"ns", "namespace", "namespaces"},
			expectedFilter: MrrFilter{Kind: "namespace"},
		},
		{
			aliases:        []string{"no", "node", "nodes"},
			expectedFilter: MrrFilter{Kind: "namespace"},
		},
	}

	for _, test := range tests {
		for _, alias := range test.aliases {
			buf.Reset()
			err := cmd.RunE(cmd, []string{alias})
			if err != nil {
				t.Errorf("Running [get %v]: got error: %v", alias, err)
			} else {
				if !reflect.DeepEqual(tc.lastFilter, test.expectedFilter) {
					t.Errorf("Running [get %v]: expected filter %v, got %v", alias, test.expectedFilter, tc.lastFilter)
				}
				if buf.String() != expectedOutput {
					t.Errorf("Running [get %v]: output [%v] was not equal to expected [%v]", alias, buf, expectedOutput)
				}
			}
		}
	}
}

func TestRunGetWithKubectlFlags(t *testing.T) {
	tc := &TestMirrorClient{}
	f := &TestFactory{mrrClient: tc}
	f.kubeconfig = Config{
		CurrentContext: "c1",
		Contexts: []ContextWrap{
			{"c1", Context{Cluster: "cluster_1", Namespace: "ns1"}},
			{"c-2", Context{Cluster: "cluster_2", Namespace: "ns2"}},
		},
		Clusters: []ClusterWrap{
			{"cluster_1", Cluster{Server: "x1.com"}},
			{"cluster_2", Cluster{Server: "x2.com"}},
			{"cluster_3", Cluster{Server: "x3.com"}},
		},
	}
	cmd := NewGetCommand(f)

	tests := []struct {
		kubectlCmd        string
		expectedNamespace string
		expectedServer    string
	}{
		{
			kubectlCmd:        "--namespace=ns-1",
			expectedNamespace: "ns-1",
		},
		{
			kubectlCmd:        "--namespace ns1",
			expectedNamespace: "ns1",
		},
		{
			kubectlCmd:        " t --namespace ns1 t --namespace=ns2 t",
			expectedNamespace: "ns2",
		},
		{
			kubectlCmd:     "--server=http://a.b:34",
			expectedServer: "http://a.b:34",
		},
		{
			kubectlCmd:     "--server s1",
			expectedServer: "s1",
		},
		{
			kubectlCmd:     "xx --server s1 xx --server=s2",
			expectedServer: "s2",
		},
		{
			kubectlCmd:        "--context=c-2",
			expectedNamespace: "ns2",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        "--context c-2",
			expectedNamespace: "ns2",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        " c --context c1 x --context c-2 c",
			expectedNamespace: "ns2",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        "--cluster=cluster_2",
			expectedNamespace: "ns1",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        "--cluster cluster_2",
			expectedNamespace: "ns1",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        "x --cluster=cluster_1 r  --cluster=cluster_2 ",
			expectedNamespace: "ns1",
			expectedServer:    "x2.com",
		},
		{
			kubectlCmd:        "--namespace=ns4 --context=c2",
			expectedNamespace: "ns4",
		},
		{
			kubectlCmd:     "--server=y1.com --cluster=cluster_2",
			expectedServer: "y1.com",
		},
		{
			kubectlCmd:     "--server=y1.com --context=c2",
			expectedServer: "y1.com",
		},
		{
			kubectlCmd:     "--cluster=cluster_3 --context=c2",
			expectedServer: "x3.com",
		},
	}

	for i, test := range tests {
		cmd.Flags().Set("kubectl-flags", test.kubectlCmd)
		cmd.RunE(cmd, []string{"po"})
		if test.expectedNamespace != "" && test.expectedNamespace != tc.lastFilter.Namespace {
			t.Errorf("Test %d: expected namespace %v, got %v", i, test.expectedNamespace, tc.lastFilter.Namespace)
		}
		if test.expectedServer != "" && test.expectedServer != tc.lastFilter.Server {
			t.Errorf("Test %d: expected server %v, got %v", i, test.expectedServer, tc.lastFilter.Server)
		}
	}
}

func TestRunGetClientError(t *testing.T) {
	tc := &TestMirrorClient{
		err: fmt.Errorf("TestFailure"),
	}
	f := &TestFactory{mrrClient: tc}
	cmd := NewGetCommand(f)

	tests := []string{"pod", "service"}
	for _, test := range tests {
		err := cmd.RunE(cmd, []string{test})
		if !strings.Contains(err.Error(), tc.err.Error()) {
			t.Errorf("Running [get %v]: error output [%v] was not equal to expected [%v]", test, err, tc.err)
		}
	}
}
