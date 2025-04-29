//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/workspace"

	"github.com/urfave/cli/v3"
)

var initCmd = cli.Command{
	DisableSliceFlagSeparator: true,
	Name:                      "init",
	Usage:                     "Initialize a new virtual machine",
	Action:                    initMachine,
	Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
		return ctx, nil
	},

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "workspace",
			Usage:    "Path to the workspace directory",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "Name of the virtual machine",
			Value:    "default",
			Required: true,
		},
		&cli.IntFlag{
			Name:  "ppid",
			Usage: "Parent process id, if not given, the ppid is the current process's ppid",
			Value: int64(os.Getpid()),
		},
		&cli.IntFlag{
			Name:  "cpus",
			Usage: "Number of CPUs to allocate to the VM",
			Value: int64(1),
		},
		&cli.IntFlag{
			Name:  "memory",
			Usage: "Amount of memory (in MB) to allocate to the VM",
			Value: 512, //nolint:mnd
		},
		&cli.StringSliceFlag{
			Name:    "volume",
			Usage:   "Host directory to mount into the VM",
			Aliases: []string{"v"},
		},
		&cli.StringFlag{
			Name:     "boot",
			Usage:    "Boot image to use for the VM",
			Aliases:  []string{"b"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "boot-version",
			Usage:    "version field the control boot image should be re-initialized, if the given version is not equal to the current version, re-initialize the boot image",
			Value:    "v1.0",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "data-version",
			Usage:    "version field the control data image should be re-initialized, if the given version is not equal to the current version, re-initialize the data image",
			Value:    "v1.0",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "vmm",
			Usage: "VMM provider to use",
			Value: "krunkit",
		},
		&cli.StringFlag{
			Name:  "report-url",
			Usage: "URL to send report events to",
		},
	},
}

func initMachine(ctx context.Context, command *cli.Command) error {
	opts := &vmconfig.VMOpts{
		VMName:      command.String("name"),
		Workspace:   command.String("workspace"),
		PPID:        command.Int("ppid"),
		CPUs:        command.Int("cpus"),
		MemoryInMiB: command.Int("memory"),
		Volumes:     command.StringSlice("volume"),
		BootImage:   command.String("boot"),
		BootVersion: command.String("boot-version"),
		DataVersion: command.String("data-version"),
	}

	// add a default mount point that store generated ignition scripts
	opts.Volumes = append(opts.Volumes, define.IgnMnt)

	workspace.SetWorkspace(opts.Workspace)

	events.ReportURL = opts.ReportURL

	vmType, err := vmconfig.GetProvider()
	if err != nil {
		return fmt.Errorf("get provider failed: %w", err)
	}
	opts.VMType = vmType

	vmcFile := filepath.Join(workspace.GetWorkspace(), define.ConfigPrefixDir, fmt.Sprintf("%s.json", opts.VMName))

	var reinit bool
	mc, err := vmconfig.LoadMachineFromFQPath(vmcFile)
	if err != nil {
		// set reinit flag to true means the vm need to be full reset
		reinit = true
	}

	if reinit {
		events.NotifyInit(events.InitNewMachine)
		mc, err = shim.Init(opts)
	} else {
		events.NotifyInit(events.InitUpdateConfig)
		mc, err = shim.Update(mc, opts)
	}

	if err != nil {
		return fmt.Errorf("init machine failed: %w", err)
	}

	if err := mc.Write(); err != nil {
		return fmt.Errorf("write machine config file failed: %w", err)
	}

	events.NotifyInit(events.InitSuccess)

	return nil
}
