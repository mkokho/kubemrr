package app

import (
	"bytes"
	"fmt"
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
			Pod{ObjectMeta: ObjectMeta{Name: "pod1"}},
			Pod{ObjectMeta: ObjectMeta{Name: "pod2"}},
		},
		services: []Service{
			Service{ObjectMeta: ObjectMeta{Name: "service1"}},
			Service{ObjectMeta: ObjectMeta{Name: "service2"}},
		},
		deployments: []Deployment{
			Deployment{ObjectMeta: ObjectMeta{Name: "deployment1"}},
			Deployment{ObjectMeta: ObjectMeta{Name: "deployment2"}},
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
