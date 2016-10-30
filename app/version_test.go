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

	if buf.String() != "0.6" {
		t.Errorf("Expected verion 0.6, got %s", buf.String())
	}
}
