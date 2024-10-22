package main

import (
	"bauklotze/cmd/bauklotze/machine"
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/network"
	"bauklotze/pkg/notifyexit"
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

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
)

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

	logLevel  = "info"
	useStdout = ""
)

func init() {
	// 初始化 rootCmd,
	// 在 rootCmd 初始化之前，子命令 如 start,stop,init 等的 cmd 已经被提前初始化
	rootCmd.SetUsageTemplate(usageTemplate)
	cobra.OnInitialize(
		loggingHook,
		stdOutHook,
	)

	lFlags := rootCmd.Flags()
	pFlags := rootCmd.PersistentFlags()

	logLevelFlagName := "log-level"
	pFlags.StringVar(&logLevel, logLevelFlagName, logLevel, fmt.Sprintf("Log messages above specified level"))

	outFlagName := "log-out"
	lFlags.StringVar(&useStdout, outFlagName, "", "Send output (stdout) from podman to a file")

	ovmHomedir := machine.Workspace
	pFlags.StringVar(&ovmHomedir, ovmHomedir, "", "Bauklotze's HOME dif, default get by $HOME")
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
}

func stdOutHook() {
	if useStdout != "" {
		if fd, err := os.OpenFile(useStdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm); err == nil {
			logrus.SetOutput(fd)
		} else {
			fmt.Fprintf(os.Stderr, "Warring: unable to open file for standard output: %s\n", err.Error())
		}
	}
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}

	if err := terminal.SetConsole(); err != nil {
		logrus.Warnf(err.Error())
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
		network.Reporter.SendEventToOvmJs("error", fmt.Sprintf("Error: %v", err))
		fmt.Fprintln(os.Stderr, formatError(err))
		registry.SetExitCode(1)
		notifyexit.NotifyExit(registry.GetExitCode())
	} else {
		registry.SetExitCode(0)
		notifyexit.NotifyExit(registry.GetExitCode())
	}
}

func loggingHook() {
	var found bool
	for _, l := range LogLevels {
		if l == strings.ToLower(logLevel) {
			found = true
			break
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Log Level %q is not supported, choose from: %s\n", logLevel, strings.Join(LogLevels, ", "))
		level, _ := logrus.ParseLevel("info")
		logrus.SetLevel(level)
		return
	}

	level, _ := logrus.ParseLevel(logLevel)
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
