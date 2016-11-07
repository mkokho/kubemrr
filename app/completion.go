package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

func NewCompletionCommand(f Factory) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "completion",
		Short: "Create completion script for kubectl (or alias)",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunAlias(f, cmd, args)
			if err != nil {
				fmt.Fprint(f.StdErr(), err.Error())
				os.Exit(1)
			}
		},
	}

	AddCommonFlags(cmd)
	cmd.Flags().String("kubectl-alias", "kubectl", "Alias of your kubectl command")
	cmd.Flags().String("kubemrr-path", "kubemrr", "Path to the kubemrr command, if it is outside $PATH variable")

	return cmd
}

func RunAlias(f Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Shell must be specified, either 'bash' or 'zsh' \n")
	}

	if len(args) > 1 {
		return fmt.Errorf("Expected exactly one argument, either 'bash' or 'zsh'")
	}

	shell := args[0]
	var in string
	switch shell {
	case "bash":
		in = bash_template
	case "zsh":
		in = zsh_template
	default:
		return fmt.Errorf("Only bash and zsh are supported, given [%v]", shell)
	}

	var err error
	c := replacement{
		kubectlAlias:   "kubectl",
		kubemrrPort:    33033,
		kubemrrAddress: "0.0.0.0",
		kubemrrPath:    "kubemrr",
	}

	if c.kubemrrPort, err = cmd.Flags().GetInt("port"); err != nil {
		return err
	}
	if c.kubemrrAddress, err = cmd.Flags().GetString("address"); err != nil {
		return err
	}
	if c.kubectlAlias, err = cmd.Flags().GetString("kubectl-alias"); err != nil {
		return err
	}
	if c.kubemrrPath, err = cmd.Flags().GetString("kubemrr-path"); err != nil {
		return err
	}

	in = fmt.Sprintf("# Below is your completion script for %s with %+v \n", shell, c) + in
	in = strings.Replace(in, "[[kubectl_alias]]", c.kubectlAlias, -1)
	in = strings.Replace(in, "[[kubemrr_path]]", c.kubemrrPath, -1)
	in = strings.Replace(in, "[[kubemrr_address]]", c.kubemrrAddress, -1)
	in = strings.Replace(in, "[[kubemrr_port]]", strconv.Itoa(c.kubemrrPort), -1)
	in = in + fmt.Sprintf("# Above is your completion script for %s with %+v \n", shell, c)

	fmt.Fprint(f.StdOut(), in)
	return nil
}

type replacement struct {
	kubectlAlias   string
	kubemrrPort    int
	kubemrrAddress string
	kubemrrPath    string
}
