package completion

import "github.com/spf13/cobra"

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
)

func AutocompleteNone(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func AutocompleteDefault(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveDefault
}

func AutocompleteLogLevel(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return LogLevels, cobra.ShellCompDirectiveNoFileComp
}
