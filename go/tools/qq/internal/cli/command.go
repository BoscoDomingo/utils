package cli

import (
	appcore "github.com/BoscoDomingo/utils/go/tools/qq/internal/app"

	"github.com/spf13/cobra"
)

const selectorFlagValue = "__qq_select_backend__"

// New builds the root Cobra command around the provided application.
func New(qqApp *appcore.App, args []string) *cobra.Command {
	rawArgs := append([]string(nil), args...)

	cmd := &cobra.Command{
		Use:           "qq [flags] [prompt]",
		Short:         "Ask a quick question with the first available AI backend",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return qqApp.Run(cmd.Context(), rawArgs)
		},
	}
	cmd.SetOut(qqApp.Stdout())
	cmd.SetErr(qqApp.Stderr())
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.ValidArgsFunction = noFileCompletion

	addBackendFlag(cmd, "backend", "b")
	addBackendFlag(cmd, "provider", "p")

	if len(args) > 0 && args[0] == "completion" {
		cmd.AddCommand(newCompletionCommand())
	}

	return cmd
}

func addBackendFlag(cmd *cobra.Command, name string, shorthand string) {
	cmd.Flags().StringP(name, shorthand, "", "backend/provider to use; omit value to choose")
	markFlagOptional(cmd, name)
	mustRegisterBackendCompletion(cmd, name)
}

func markFlagOptional(cmd *cobra.Command, name string) {
	flag := cmd.Flags().Lookup(name)
	if flag != nil {
		flag.NoOptDefVal = selectorFlagValue
	}
}

func noFileCompletion(
	_ *cobra.Command,
	_ []string,
	_ string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
