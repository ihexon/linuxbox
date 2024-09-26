package main

import (
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/completion"
	"bauklotze/pkg/terminal"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

const helpTemplate = `{{.Short}}

Description:
  {{.Long}}

{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

const usageTemplate = `Usage:{{if (and .Runnable (not .HasAvailableSubCommands))}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.UseLine}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
{{end}}
`

func flagErrorFunc(c *cobra.Command, e error) error {
	e = fmt.Errorf("%w\nSee '%s --help'", e, c.CommandPath())
	return e
}

var (
	rootCmd = &cobra.Command{
		Use:                   filepath.Base(os.Args[0]) + " [options]",
		Long:                  "Manage your bugbox",
		SilenceUsage:          true,
		SilenceErrors:         true,
		TraverseChildren:      true,
		PersistentPreRunE:     persistentPreRunE,
		RunE:                  validata.SubCommandExists,
		PersistentPostRunE:    persistentPostRunE,
		DisableFlagsInUseLine: true,
	}

	logLevel = "info"
)

func init() {
	// 初始化 rootCmd,
	// 在 rootCmd 初始化之前，子命令 如 start,stop,init 等的 cmd 已经被提前初始化
	rootCmd.SetUsageTemplate(usageTemplate)
	cobra.OnInitialize(
		loggingHook,
	)

	pFlags := rootCmd.PersistentFlags()

	logLevelFlagName := "log-level"
	pFlags.StringVar(&logLevel, logLevelFlagName, logLevel, fmt.Sprintf("Log messages above specified level (%s)", strings.Join(completion.LogLevels, ", ")))
	_ = rootCmd.RegisterFlagCompletionFunc(logLevelFlagName, completion.AutocompleteLogLevel)

	ovmHomedir := "workdir"
	pFlags.StringVar(&ovmHomedir, ovmHomedir, "", "Bauklotze's HOME dif, default get by $HOME")
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
	os.Exit(0)
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}

	if err := terminal.SetConsole(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	rootCmd.SetFlagErrorFunc(flagErrorFunc)
	return rootCmd
}

func addCommand(c registry.CliCommand) {
	parent := rootCmd
	if c.Parent != nil {
		parent = c.Parent
	}
	parent.AddCommand(c.Command)
	c.Command.SetFlagErrorFunc(flagErrorFunc)
	c.Command.SetHelpTemplate(helpTemplate)
	c.Command.SetUsageTemplate(usageTemplate)
	c.Command.DisableFlagsInUseLine = true
}

func RootCmdExecute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, formatError(err))
	}
	os.Exit(registry.GetExitCode())
}

func loggingHook() {
	var found bool
	for _, l := range completion.LogLevels {
		if l == strings.ToLower(logLevel) {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Log Level %q is not supported, choose from: %s\n", logLevel, strings.Join(completion.LogLevels, ", "))
		os.Exit(1)
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	logrus.SetLevel(level)

	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		logrus.Infof("%s filtering at log level %s", os.Args[0], logrus.GetLevel())
	}
}

func formatError(err error) string {
	var message string
	switch {
	default:
		if logrus.IsLevelEnabled(logrus.TraceLevel) {
			message = fmt.Sprintf("Error: %+v", err)
		} else {
			message = fmt.Sprintf("Error: %v", err)
		}
	}
	return message
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	logrus.Debugf("Called %s.PersistentPreRunE(%s)", cmd.Name(), strings.Join(os.Args, " "))

	if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.HasSubCommands() {
		return nil
	}
	return nil
}

func persistentPostRunE(cmd *cobra.Command, args []string) error {
	return nil
}
