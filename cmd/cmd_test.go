package cmd

import (
  "testing"
	"time"
)

func TestCommands(t *testing.T) {
	go watchCmd.Run(RootCmd, []string{})
	d, _ := time.ParseDuration("100ms")
	time.Sleep(d)
  getCmd.Run(RootCmd, []string{})
}
