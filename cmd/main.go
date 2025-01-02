//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"bauklotze/pkg/machine/events"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cmdflags "bauklotze/cmd/bauklotze/flags"
	_ "bauklotze/cmd/bauklotze/machine"
	"bauklotze/cmd/bauklotze/validata"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/terminal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		Use:              filepath.Base(os.Args[0]) + " [options]",
		Long:             "Manage your bugbox",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		// PersistentPreRunE/PreRunE/RunE will run after rootCmd.ExecuteContext(context.Background()), also run after init()
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("==========================================================")
			logrus.Infof("OVM VERSION dev-%s\n", define.GitCommit)
			logrus.Infof("FULL OVM COMMANDLINE: %v\n", os.Args)
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("WORKSPACE: %s", homeDir)
			return nil
		},
		PostRunE:              validata.SubCommandExists,
		DisableFlagsInUseLine: true,
	}

	commonOpts = &define.CommonOptions{}
	logLevel   = ""
	logOut     = ""
	homeDir    = ""
)

func init() {
	cobra.OnInitialize(
		loggingHook,
		stdOutHook,
		ReportHook,
	)
	cobra.EnableTraverseRunHooks = true
	rootCmd.SetUsageTemplate(usageTemplate)
	pFlags := rootCmd.PersistentFlags()

	logLevelFlagName := cmdflags.LogLevelFlag
	pFlags.StringVar(&logLevel, logLevelFlagName, cmdflags.DefaultLogLevel, "Log messages above specified level,by default is info")

	outFlagName := cmdflags.LogOutFlag
	pFlags.StringVar(&logOut, outFlagName, cmdflags.FileBased, "If set --log-out console, send output to terminal, if set --log-out file, send output to ${workspace}/logs/ovm.log")

	ovmHomedir := cmdflags.WorkspaceFlag
	pFlags.StringVar(&homeDir, ovmHomedir, "", "Bauklotze's HOME directory, this flag is mandatory required")
	_ = rootCmd.MarkPersistentFlagRequired(ovmHomedir)

	ReportURLFlag := cmdflags.ReportURLFlag
	pFlags.StringVar(&commonOpts.ReportURL, ReportURLFlag, "", "Report events to the url")

	ppidFlagName := cmdflags.PpidFlag
	defaultPPID, _ := system.GetPPID(int32(os.Getpid()))
	pFlags.Int32Var(&commonOpts.PPID, ppidFlagName, defaultPPID, "Parent process id, if not given, the ppid is the current process's ppid")
}

func main() {
	rootCmd = parseCommands()
	RootCmdExecute()
}

func ReportHook() {
	if commonOpts.ReportURL != "" {
		logrus.Infof("ReportHook(): Report events to the url: %s\n", commonOpts.ReportURL)
		events.ReportURL = commonOpts.ReportURL
	} else {
		logrus.Infof("No report url provided, skip report events\n")
	}
}

const defaultLogFileSize = 10 * 1024 * 1024 // 10MB
const defaultSeekSize = 5 * 1024 * 1024     // 5MB

func stdOutHook() {
	_ = os.Stdin.Close()
	_logOut, _ := rootCmd.PersistentFlags().GetString(cmdflags.LogOutFlag)
	// --log-out must use with --workspace
	hasWorkSpace, _ := rootCmd.PersistentFlags().GetString(cmdflags.WorkspaceFlag)
	if hasWorkSpace == "" {
		return
	}

	if _logOut == cmdflags.FileBased {
		logFile := filepath.Join(homeDir, "logs", "ovm.log")
		err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to create directory for log file: %s\n", err.Error())
		}

		logrus.Infof("Log all output to file %s\n", logFile)

		// discard first 5MB if the logfile larger than 10MB
		fileInfo, err := os.Stat(logFile)
		if err == nil {
			if fileInfo.Size() <= defaultLogFileSize { // 10MB
				logrus.Infof("File size is within limit, no changes made.")
			} else {
				logrus.Infof("File size is %d bytes, trimming the file.", fileInfo.Size())
				// If the logFile large then 10*1024*1024 (10MB)
				file, _ := os.Open(logFile)
				defer file.Close()
				_, _ = file.Seek(defaultSeekSize, io.SeekStart) // 5MB
				tempFile, _ := os.CreateTemp("", "trimmed-ovm-log.txt")
				defer tempFile.Close()
				_, _ = io.Copy(tempFile, file)
				file.Close()
				tempFile.Close()
				_ = os.Rename(tempFile.Name(), logFile)
				logrus.Infof("Successfully trimmed the file.")
			}
		}

		fd, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to open file for standard output: %s\n", err.Error())
		} else {
			os.Stdout = fd
			os.Stderr = fd
			logrus.SetOutput(fd)
		}
	} else {
		logrus.Infof("Log all output to console\n")
	}
}

func parseCommands() *cobra.Command {
	for _, c := range registry.Commands {
		addCommand(c)
	}

	if err := terminal.SetConsole(); err != nil {
		logrus.Warn(err.Error())
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

func formatError(err error) string {
	message := fmt.Sprintf("Error: %+v", err)
	return message
}

func RootCmdExecute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, formatError(err))
		registry.SetExitCode(1)
	} else {
		registry.SetExitCode(0)
	}
	events.NotifyExit()
	os.Exit(registry.GetExitCode())
}

func loggingHook() {
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Log Level %q is not supported, choose from: %s\n", logLevel, strings.Join(LogLevels, ", "))
		level, _ = logrus.ParseLevel("error")
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
}
