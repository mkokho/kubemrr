package app

import (
	"bytes"
	"testing"
)

func TestRunVersion(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	f := &TestFactory{stdOut: buf}
	cmd := NewVersionCommand(f)
	cmd.Run(cmd, []string{})

	expected := "kubemrr-0.6.0"
	if buf.String() != expected {
		t.Errorf("Expected verion %s, got %s", expected, buf.String())
	}
}
