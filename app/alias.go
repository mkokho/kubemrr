package app

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

func NewAliasCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "alias",
		Short: "Create completion script for kubectl (or alias)",
		Run: func(cmd *cobra.Command, args []string) {
			RunAlias(f, cmd, args)
		},
	}
	AddCommonFlags(cmd)

	return cmd
}

func RunAlias(f Factory, cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(f.StdErr(), "You must specify the alias")
		return
	}

	if len(args) > 1 {
		fmt.Fprintf(f.StdErr(), "Only one argument is expected")
		return
	}

	var err error
	c := replacement{
		kubectlAlias:   args[0],
		kubemrrPort:    33033,
		kubemrrAddress: "0.0.0.0",
		kubemrrAlias:   "kubemrr",
	}

	if c.kubemrrPort, err = cmd.Flags().GetInt("port"); err != nil {
		log.Fatal(err)
	}
	if c.kubemrrAddress, err = cmd.Flags().GetString("address"); err != nil {
		log.Fatal(err)
	}

	in := bash_template
	in = strings.Replace(in, "[[kubectl_alias]]", c.kubectlAlias, -1)
	in = strings.Replace(in, "[[kubemrr_alias]]", c.kubemrrAlias, -1)
	in = strings.Replace(in, "[[kubemrr_address]]", c.kubemrrAddress, -1)
	in = strings.Replace(in, "[[kubemrr_port]]", strconv.Itoa(c.kubemrrPort), -1)

	fmt.Fprint(f.StdOut(), in)
}

type replacement struct {
	kubectlAlias   string
	kubemrrPort    int
	kubemrrAddress string
	kubemrrAlias   string
}
