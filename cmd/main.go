package main

import (
	cmdflags "bauklotze/cmd/bauklotze/flags"
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	machine2 "bauklotze/pkg/machine"
	"bauklotze/pkg/machine/system"
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
		RunE:                  validata.SubCommandExists,
		DisableFlagsInUseLine: true,
	}
	logLevel = ""
	logOut   = ""
	homeDir  = ""
)

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)
	lFlags := rootCmd.Flags()
	pFlags := rootCmd.PersistentFlags()

	logLevelFlagName := cmdflags.LogLevelFlag
	pFlags.StringVar(&logLevel, logLevelFlagName, cmdflags.LogLevel, fmt.Sprintf("Log messages above specified level"))

	outFlagName := cmdflags.LogOutFlag
	lFlags.StringVar(&logOut, outFlagName, cmdflags.FileBased, "If set --log-out console, send output to terminal, if set --log-out file, send output to ${workspace}/logs/")

	ovmHomedir := cmdflags.WorkspaceFlag
	pFlags.StringVar(&homeDir, ovmHomedir, "", "Bauklotze's HOME dif, default get by $HOME")

	cobra.OnInitialize(
		loggingHook,
		stdOutHook,
	)
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
}

func stdOutHook() {
	if rootCmd.Flag(cmdflags.LogOutFlag).Value.String() == cmdflags.FileBased {
		// ${workspace}/logs
		err := os.MkdirAll(filepath.Join(rootCmd.Flag(cmdflags.WorkspaceFlag).Value.String(), "logs"), os.ModePerm)
		if err != nil {
			logrus.Errorf("Unable to create directory for log file: %s", err.Error())
			return
		}
		// ${workspace}/logs/ovm.log
		logfile := filepath.Join(rootCmd.Flag(cmdflags.WorkspaceFlag).Value.String(), "logs", "ovm.log")
		if fd, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm); err == nil {
			logrus.SetOutput(fd)
		} else {
			logrus.Errorf("Warring: unable to open file for standard output: %s\n", err.Error())
			return
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
	// Make sure always kill the `gvproxy` and `krunkit` process
	_ = system.KillProcess(machine2.GlobalPIDs.GetGvproxyPID())
	_ = system.KillProcess(machine2.GlobalPIDs.GetKrunkitPID())
	if err != nil {
		logrus.Errorf(err.Error())
		network.Reporter.SendEventToOvmJs("error", fmt.Sprintf("Error: %v", err))
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
