package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"envchanger/internal/envpath"
)

type usageError struct {
	cmd     *cobra.Command
	message string
}

func (e *usageError) Error() string {
	return e.message
}

func newUsageError(cmd *cobra.Command, message string) error {
	return &usageError{cmd: cmd, message: message}
}

func main() {
	rootCmd := newRootCommand()

	if err := rootCmd.Execute(); err != nil {
		var uErr *usageError
		if errors.As(err, &uErr) {
			uErr.cmd.SetErr(os.Stderr)
			uErr.cmd.SetOut(os.Stderr)
			if uErr.message != "" {
				uErr.cmd.PrintErrln(uErr.message)
			}
			_ = uErr.cmd.Usage()
			os.Exit(2)
		}
		exitWithError(err)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changeenv [target-env]",
		Short: "Switch to the same relative directory in another environment tree.",
		Long: `Switch the current working directory to the same relative location in another environment tree.

Examples:
  changeenv prod

To change shell directories directly, wrap with a shell function:
  cenv() { cd "$(changeenv "$1")"; }
`,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return newUsageError(cmd, "target environment argument is required")
			}

			targetEnv := strings.TrimSpace(args[0])
			if targetEnv == "" {
				return newUsageError(cmd, "target environment must not be empty")
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("determine current directory: %w", err)
			}

			targetPath, err := envpath.Switch(cwd, targetEnv)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), targetPath)
			return nil
		},
	}

	cmd.AddCommand(newConfigureCommand())

	return cmd
}

func newConfigureCommand() *cobra.Command {
	var autoApply bool

	cmd := &cobra.Command{
		Use:           "configure",
		Short:         "Print or append the shell helper function.",
		Long:          `Print the helper shell function and PATH export snippet, or append them directly to your shell configuration file.`,
		Args:          validateNoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigure(autoApply)
		},
	}

	cmd.Flags().BoolVar(&autoApply, "create", false, "append the shell helper to the detected configuration file when running configure")

	return cmd
}

func validateNoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return newUsageError(cmd, "configure does not accept positional arguments")
	}
	return nil
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "changeenv: %v\n", err)
	os.Exit(1)
}
