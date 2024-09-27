package completion

import "github.com/spf13/cobra"

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
)

// AutocompleteNone - Block the default shell completion (no paths)
func AutocompleteNone(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
