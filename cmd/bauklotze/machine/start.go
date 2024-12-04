//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmdflags "bauklotze/cmd/bauklotze/flags"
	"bauklotze/cmd/registry"
	"bauklotze/pkg/api/server"
	cmdproxy "bauklotze/pkg/cliproxy"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	system2 "bauklotze/pkg/system"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	startCmd = &cobra.Command{
		Use:   "start [options] [MACHINE]",
		Short: "Start an existing machine",
		Long:  "Start a managed virtual machine ",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return machinePreRunE(cmd, args)
		}, // Get Provider and set workdir if needed
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd, args)
		},
		Args:    cobra.MaximumNArgs(1),
		Example: `bauklotze machine start`,
	}
	startOpts = define.StartOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})
}

func start(cmd *cobra.Command, args []string) error {
	// Killall ovm process before running ovm, this should never happen,
	// but we still do this to avoid any issue
	pids := []int32{}
	pidskrun, _ := system2.FindPIDByCmdline(".oomol-studio/ovm-krun/data/libkrun/default-arm64.raw")
	pidsGvp, _ := system2.FindPIDByCmdline(".oomol-studio/ovm-krun/tmp/gvproxy.pid")

	pids = append(pids, pidskrun...)
	pids = append(pids, pidsGvp...)

	for _, pid := range pids {
		logrus.Warnf("Killing PID: %d", pid)
		_ = system.KillProcess(int(pid))
	}

	ppid, _ := cmd.Flags().GetInt32(cmdflags.PpidFlag) // Get PPID from --ppid flag
	logrus.Infof("PID is [%d], PPID is: %d", os.Getpid(), ppid)
	reportURL := cmd.Flag(cmdflags.ReportUrlFlag).Value.String()

	startOpts.CommonOptions.ReportUrl = reportURL
	startOpts.CommonOptions.PPID = ppid

	// now we have dirs, and we do not need env.GetMachineDirs again
	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return err
	}

	logrus.Infof("ConfigDir:     %s", dirs.ConfigDir.GetPath())
	logrus.Infof("DataDir:       %s", dirs.DataDir.GetPath())
	logrus.Infof("ImageCacheDir: %s", dirs.ImageCacheDir.GetPath())
	logrus.Infof("RuntimeDir:    %s", dirs.RuntimeDir.GetPath())
	logrus.Infof("LogsDir:       %s", dirs.LogsDir.GetPath())

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigChan:
			// Listen SIGTERM signal, and return an error
			// when the signal is received
			return fmt.Errorf("Received shutdown signal")
		}
	})

	g.Go(func() error {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				isRunning, err := system.IsProcesSAlive([]int32{ppid})
				if !isRunning {
					return fmt.Errorf("check PPID running: %v", err)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	var mc *vmconfigs.MachineConfig

	g.Go(func() error {
		errCh := make(chan error, 1)
		vmName := define.DefaultMachineName
		if len(args) > 0 && len(args[0]) > 0 {
			vmName = args[0]
		}
		mc, err = vmconfigs.LoadMachineByName(vmName, dirs)
		if err != nil {
			return err
		}
		logrus.Infof("Starting machine %q\n", vmName)
		go func() {
			network.Reporter.SendEventToOvmJs("start", fmt.Sprintf("Starting machine: %s ....", vmName))
			// Note: shim.Start is no block function
			errCh <- shim.Start(ctx, mc, provider, dirs, startOpts)
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err = <-errCh:
			if err == nil {
				logrus.Infof("Machine %q started successfully\n", vmName)
			}
		}
		return err
	})

	// Start a goroutine running api service
	g.Go(func() error {
		listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"
		logrus.Infof("Starting API Server in %s\n", listenPath)
		apiURL, _ := url.Parse(listenPath)
		return server.RestService(ctx, apiURL) // server.RestService must now subscribe to ctx
	})

	go func() {
		logrus.Infof("ovm sshd starting...")
		cmdProxyErr := cmdproxy.RunCMDProxy()
		if cmdProxyErr != nil {
			logrus.Errorf("ovm sshd running failed, %v", cmdProxyErr)
		}
	}()

	err = g.Wait()

	if mc != nil {
		logrus.Infof("Do sync in virtualMachine....")
		if syncErr := machine.CommonSSHSilent(mc.SSH.RemoteUsername, mc.SSH.IdentityPath, mc.Name, mc.SSH.Port, []string{"sync"}); syncErr != nil {
			logrus.Warnf("Sync failed: %v", syncErr)
		}
	}

	if kruncmd := machine.GlobalCmds.GetKrunCmd(); kruncmd != nil {
		logrus.Infof("--> Killing krun PID: %d", kruncmd.Process.Pid)
		_ = kruncmd.Process.Kill()
		_ = kruncmd.Wait()
	}

	if gvcmd := machine.GlobalCmds.GetGvproxyCmd(); gvcmd != nil {
		logrus.Infof("--> Killing gvproxy PID: %d", gvcmd.Process.Pid)
		_ = gvcmd.Process.Kill()
		_ = gvcmd.Wait()
	}

	return err
}
