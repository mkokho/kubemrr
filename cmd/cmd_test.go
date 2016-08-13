package cmd

import (
  "testing"
	"time"
	"github.com/spf13/cobra"
	"github.com/mkokho/kubemrr/pkg"
)

func TestCommands(t *testing.T) {
	cmd := &cobra.Command{}
	pkg.AddCommonFlags(cmd)
	go watchCmd.Run(cmd, []string{})
	d, _ := time.ParseDuration("100ms")
	time.Sleep(d)
  getCmd.Run(cmd, []string{})
}
