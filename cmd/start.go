//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/machine/krunkit"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vfkit"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/workspace"
	"bauklotze/pkg/registry"
	"bauklotze/pkg/system"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

var startCmd = cli.Command{
	Name:  "start",
	Usage: "Start a virtual machine",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "workspace",
			Usage:    "workspace directory",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "machine name",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "ppid",
			Usage:    "process parent pid file",
			Required: false,
			Value:    1,
		},
	},
	Action: start,
}

const tickerInterval = 300 * time.Millisecond

func start(parentCtx context.Context, command *cli.Command) error {
	opts := &vmconfig.VMOpts{
		VMName:    command.String("name"),
		PPID:      command.Int("ppid"),
		Workspace: command.String("workspace"),
	}

	workspace.SetWorkspace(opts.Workspace)
	events.ReportURL = opts.ReportURL

	vmType, err := vmconfig.GetProvider()
	if err != nil {
		return fmt.Errorf("get provider failed: %w", err)
	}
	opts.VMType = vmType

	vmcFile := filepath.Join(workspace.GetWorkspace(), define.ConfigPrefixDir, fmt.Sprintf("%s.json", opts.VMName))

	// We first check the status of the pid passed in via --ppid,
	// and if it is inactive, exit immediately without running any of the following code
	isRunning, err := system.IsProcessAliveV4(int(opts.PPID))
	if !isRunning {
		return fmt.Errorf("PPID %d exited, possible error: %w", opts.PPID, err)
	}

	events.NotifyRun(events.LoadMachineConfig)
	mc, err := vmconfig.LoadMachineFromFQPath(vmcFile)
	if err != nil {
		return fmt.Errorf("load machine config file failed: %w", err)
	}

	g, ctx := errgroup.WithContext(parentCtx)
	WatchPPID(g, ctx, opts)
	ListenSignal(g, ctx)
	StartRestAPI(g, ctx, mc)
	startMachine(g, ctx, mc, opts)

	return g.Wait() //nolint:wrapcheck
}

func startMachine(g *errgroup.Group, ctx context.Context, mc *vmconfig.MachineConfig, opts *vmconfig.VMOpts) {
	g.Go(func() error {
		var vmp vmconfig.VMProvider
		switch opts.VMType {
		case vmconfig.LibKrun:
			logrus.Infof("vm provider: %q", opts.VMType.String())
			vmp = krunkit.NewProvider()
		case vmconfig.VFkit:
			vmp = vfkit.NewProvider()
		default:
			return fmt.Errorf("invalid vmm type")
		}

		if err := shim.Start(ctx, mc, vmp); err != nil {
			return fmt.Errorf("start machine %q error: %w", mc.VMName, err)
		}

		// NOTE:
		// shim.Wait do not need to support context.
		// once the parent context is done, the cmd will be killed by the os.Process.Kill(). so the shim.Wait will return immediately.
		return shim.Wait(registry.GetCmds()...)
	})
}

func StartRestAPI(g *errgroup.Group, ctx context.Context, mc *vmconfig.MachineConfig) {
	g.Go(func() error {
		endPoint := filepath.Join(filepath.Dir(mc.Dirs.SocksDir), define.RESTAPIEndpointName)
		logrus.Infof("Start rest api service at %q", endPoint)
		return server.RestService(ctx, mc, endPoint)
	})
}

func ListenSignal(g *errgroup.Group, ctx context.Context) {
	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sigChan:
			return fmt.Errorf("catch signal: %v", s.String())
		}
	})
}

func WatchPPID(g *errgroup.Group, ctx context.Context, opts *vmconfig.VMOpts) {
	g.Go(func() error {
		ticker := time.NewTicker(tickerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				isRunning, err := system.IsProcessAliveV4(int(opts.PPID))
				if !isRunning {
					return fmt.Errorf("PPID %d exited, possible error: %w", opts.PPID, err)
				}
			}
		}
	})
}
