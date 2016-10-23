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
			output: "specify the resource",
		},
		{
			args:   []string{"1", "2"},
			output: "one argument",
		},
		{
			args:   []string{"k8s-resource"},
			output: "Unsupported resource type",
		},
	}

	buf := bytes.NewBuffer([]byte{})
	f := &TestFactory{stdErr: buf}
	cmd := NewGetCommand(f)

	for i, test := range tests {
		buf.Reset()
		cmd.Run(cmd, test.args)
		if buf.Len() == 0 {
			t.Errorf("Test %d: nothing has been written to the error output, expected: %v", i, test.output)
		}

		if !strings.Contains(buf.String(), test.output) {
			t.Errorf("Test %d: output [%v] does not contains expected [%v]", i, buf, test.output)
		}
	}
}

func TestRunGet(t *testing.T) {
	tc := &TestMirrorClient{
		pods: []Pod{
			{ObjectMeta: ObjectMeta{Name: "pod1"}},
			{ObjectMeta: ObjectMeta{Name: "pod2"}},
		},
		services: []Service{
			{ObjectMeta: ObjectMeta{Name: "service1"}},
			{ObjectMeta: ObjectMeta{Name: "service2"}},
		},
		deployments: []Deployment{
			{ObjectMeta: ObjectMeta{Name: "deployment1"}},
			{ObjectMeta: ObjectMeta{Name: "deployment2"}},
		},
	}
	buf := bytes.NewBuffer([]byte{})
	f := &TestFactory{mrrClient: tc, stdOut: buf}
	cmd := NewGetCommand(f)

	tests := []struct {
		aliases []string
		output  string
	}{
		{
			aliases: []string{"po", "pod", "pods"},
			output:  "pod1 pod2",
		},
		{
			aliases: []string{"svc", "service", "services"},
			output:  "service1 service2",
		},
		{
			aliases: []string{"deployment", "deployments"},
			output:  "deployment1 deployment2",
		},
	}

	for _, test := range tests {
		for _, alias := range test.aliases {
			buf.Reset()
			cmd.Run(cmd, []string{alias})
			if buf.String() != test.output {
				t.Errorf("Running [get %v]: output [%v] was not equal to expected [%v]", alias, buf, test.output)
			}
		}
	}
}

func TestRunGetClientError(t *testing.T) {
	tc := &TestMirrorClient{
		err: fmt.Errorf("TestFailure"),
	}
	buf := bytes.NewBuffer([]byte{})
	f := &TestFactory{mrrClient: tc, stdErr: buf}
	cmd := NewGetCommand(f)

	tests := []string{"pod", "service"}
	for _, test := range tests {
		buf.Reset()
		cmd.Run(cmd, []string{test})
		if !strings.Contains(buf.String(), tc.err.Error()) {
			t.Errorf("Running [get %v]: error output [%v] was not equal to expected [%v]", test, buf, tc.err)
		}
	}
}

func TestParseKubeConfigFailures(t *testing.T) {
	tests := []struct {
		filename string
		complain string
	}{
		{
			filename: "test_data/kubeconfig_missing",
			complain: "not read",
		},
		{
			filename: "test_data/kubeconfig_invalid",
			complain: "invalid",
		},
	}

	for _, test := range tests {
		_, err := parseKubeConfig(test.filename)
		if err == nil {
			t.Errorf("Expected an error for file %s", test.filename)
			continue
		}

		if !strings.Contains(err.Error(), test.complain) {
			t.Errorf("Error [%s] does not contain [%s]", err, test.complain)
		}
	}
}

func TestParseKubeConfig(t *testing.T) {
	actual, err := parseKubeConfig("test_data/kubeconfig_valid")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	expected := Config{
		CurrentContext: "prod",
		Contexts: []ContextWrap{
			{"dev", Context{"cluster_2", "red"}},
			{"prod", Context{"cluster_1", "blue"}},
		},
		Clusters: []ClusterWrap{
			{"cluster_1", Cluster{"https://foo.com"}},
			{"cluster_2", Cluster{"https://bar.com"}},
		},
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}

}
