//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package registry

import (
	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/channel"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/system"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CliCommand struct {
	Command *cobra.Command
	Parent  *cobra.Command
}

var (
	exitCode = 0
	// Commands All commands will be registin here
	Commands []CliCommand
)

func SetExitCode(code int) {
	exitCode = code
}

func GetExitCode() int {
	return exitCode
}

func NotifyAndExit(code int) {
	events.NotifyExit()
	channel.Close()
	SetExitCode(code)
	os.Exit(GetExitCode())
}

var (
	vmp  vmconfig.VMProvider
	once sync.Once
)

func initializeProvider() (vmconfig.VMProvider, error) {
	var err error
	once.Do(func() {
		vmp, err = provider.Get()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current hypervisor provider: %w", err)
	}
	return vmp, nil
}

func GetProvider() (vmconfig.VMProvider, error) {
	if vmp != nil {
		return vmp, nil
	}
	return nil, fmt.Errorf("vm provider is nil, maybe provider not initialize")
}

func showLogHeader() {
	logrus.Infof("%s", system.Version())
	logrus.Info(fmt.Sprintf("CMDLINE: %q", os.Args))
	if define.GitCommit == "" {
		define.GitCommit = "unknown"
	}
	logrus.Info(fmt.Sprintf("OVM VERSION: %s", define.GitCommit))
	logrus.Infof("OVM PID: %d, PPID: %d", os.Getpid(), allFlag.PPID)
	logrus.Infof("OVM WORKSPACE: %s", allFlag.WorkSpace)
}

func initVMProvider() (vmconfig.VMProvider, error) {
	p, err := initializeProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}
	return p, nil
}

func PreRunE(cmd *cobra.Command, args []string) error {
	err := setWorkSpace()
	if err != nil {
		return fmt.Errorf("set workspace error: %w", err)
	}
	err = redirectOutput(cmd)
	if err != nil {
		return fmt.Errorf("set logger error: %w", err)
	}
	showLogHeader()
	logrus.Infof("Try to get current hypervisor provider...")
	mp, err := initVMProvider()
	if err != nil {
		return fmt.Errorf("initialize VM provider error: %w", err)
	}
	logrus.Infof("VM Provider: %s", mp.VMType().String())

	if len(args) > 0 && args[0] != "" {
		allFlag.VMName = args[0]
	} else {
		allFlag.VMName = define.DefaultMachineName
	}
	return nil
}

func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	// Get the subcommand's subcommand from cmd
	currentCommand := cmd.Name()
	if currentCommand == "init" {
		events.CurrentStage = events.Init
	}
	if currentCommand == "start" {
		events.CurrentStage = events.Run
	}
	logrus.Infof("Current stage is %s", events.Run)
	return nil
}

const defaultLogFileSizeInMB = 10

func redirectOutput(cmd *cobra.Command) error {
	if err := os.Stdin.Close(); err != nil {
		return fmt.Errorf("unable to close stdin: %w", err)
	}

	logOut := cmd.Flags().Lookup(LogOutFlag).Value.String()
	if logOut == ConsoleBased {
		logrus.Infof("Log all output to console")
		return nil
	}

	// Get workspace
	workspace, err := vmconfig.GetWorkSpace()
	if err != nil {
		return fmt.Errorf("unable to get workspace: %w", err)
	}

	myLogFile, err := io.NewMachineFile(filepath.Join(workspace.GetPath(), define.LogPrefixDir, define.LogFileName))
	if err != nil {
		return fmt.Errorf("unable to create log file: %w", err)
	}
	// If myLogFile exist and shrink the log file
	if myLogFile.Exist() {
		if err = myLogFile.DiscardBytesAtBegin(defaultLogFileSizeInMB); err != nil {
			return fmt.Errorf("unable to shrink log file: %w", err)
		}
		// reopen the log file
		fd, err := os.OpenFile(myLogFile.GetPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open log file: %w", err)
		}
		// Set logrus output to the file
		os.Stdout = fd
		os.Stderr = fd
		logrus.SetOutput(fd)
	} else {
		// If myLogFile not exist, create a new log file
		if err = myLogFile.MakeBaseDir(); err != nil {
			return fmt.Errorf("unable to create dir: %s, %w", myLogFile.GetPath(), err)
		}
		// reopen the log file
		fd, err := os.OpenFile(myLogFile.GetPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open log file: %w", err)
		}
		os.Stdout = fd
		os.Stderr = fd
		logrus.SetOutput(fd)
	}
	return nil
}

func setWorkSpace() error {
	_, err := vmconfig.SetWorkSpace(allFlag.WorkSpace)
	if err != nil {
		return fmt.Errorf("failed to set workspace: %w", err)
	}
	return nil
}
