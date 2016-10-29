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
			aliases: []string{"po", "pod", "pods"},
			expectedFilter:  MrrFilter{Kind: "pod"},
		},
		{
			aliases: []string{"svc", "service", "services"},
			expectedFilter:  MrrFilter{Kind: "service"},
		},
		{
			aliases: []string{"deployment", "deployments"},
			expectedFilter:  MrrFilter{Kind: "deployment"},
		},
	}

	for _, test := range tests {
		for _, alias := range test.aliases {
			buf.Reset()
			cmd.Run(cmd, []string{alias})
			if !reflect.DeepEqual(tc.lastFilter, test.expectedFilter) {
				t.Errorf("Running [get %v]: expected filter %v, got %v", alias, test.expectedFilter, tc.lastFilter)
			}
			if buf.String() != expectedOutput {
				t.Errorf("Running [get %v]: output [%v] was not equal to expected [%v]", alias, buf, expectedOutput)
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
