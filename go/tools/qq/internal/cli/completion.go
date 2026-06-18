package cli

import (
	"fmt"

	"github.com/BoscoDomingo/utils/go/tools/qq/internal/backend"

	"github.com/spf13/cobra"
)

func mustRegisterBackendCompletion(cmd *cobra.Command, flagName string) {
	err := cmd.RegisterFlagCompletionFunc(flagName, func(
		_ *cobra.Command,
		_ []string,
		_ string,
	) ([]cobra.Completion, cobra.ShellCompDirective) {
		backends := backend.Supported()
		completions := make([]cobra.Completion, len(backends))
		for index, item := range backends {
			completions[index] = cobra.CompletionWithDesc(item.Name(), "qq backend")
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		panic(err)
	}
}

func mustRegisterNoFileFlagCompletion(cmd *cobra.Command, flagName string) {
	err := cmd.RegisterFlagCompletionFunc(flagName, noFileCompletion)
	if err != nil {
		panic(err)
	}
}

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}
