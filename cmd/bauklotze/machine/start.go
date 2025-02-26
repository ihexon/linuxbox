//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine"
	allFlag "bauklotze/pkg/machine/allflag"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfig"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine/events"
	"bauklotze/pkg/system"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: registry.PersistentPreRunE,
		PreRunE:           registry.PreRunE,
		RunE:              start,
		Args:              cobra.MaximumNArgs(1),
		Example:           `machine start`,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})

	lFlags := startCmd.Flags()
	ppidFlagName := registry.PpidFlag
	defaultPPID, _ := system.GetPPID(int32(os.Getpid()))
	lFlags.Int32Var(&allFlag.PPID, ppidFlagName, defaultPPID, "Parent process id, if not given, the ppid is the current process's ppid")
}

const tickerInterval = 300 * time.Millisecond

// TODO: Http Proxy
func start(cmd *cobra.Command, args []string) error {
	logrus.Infof("===================START===================")
	vmp, err := registry.GetProvider()
	if err != nil {
		return fmt.Errorf("failed to get current hypervisor provider:%w", err)
	}

	events.NotifyRun(events.LoadMachineConfig)
	mc, err := shim.GetVMConf(allFlag.VMName, []vmconfig.VMProvider{vmp})
	if err != nil {
		return fmt.Errorf("failed to get machine config: %w", err)
	}
	mc.VMProvider = vmp

	// If the user given the report url, then overwrite the report url into mc
	if allFlag.ReportURL != "" {
		mc.ReportURL = &io.VMFile{Path: allFlag.ReportURL}
	}

	g, ctx := errgroup.WithContext(context.Background())
	// We first check the status of the pid passed in via --ppid,
	// and if it is inactive, exit immediately without running any of the following code
	logrus.Infof("Check the status of the parent process: %d", allFlag.PPID)
	isRunning, err := system.IsProcessAliveV4(int(allFlag.PPID))
	if !isRunning {
		return fmt.Errorf("PPID[ %d ] exited, possible error: %w", allFlag.PPID, err)
	}

	// Check if the process of the PID passed in via --ppid is active.
	// If the PID status is inactive, an error is returned immediately.
	g.Go(func() error {
		ticker := time.NewTicker(tickerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				isRunning, err := system.IsProcessAliveV4(int(allFlag.PPID))
				if !isRunning {
					return fmt.Errorf("PPID[%d] exited, possible error: %w", allFlag.PPID, err)
				}
			}
		}
	})

	// Listen the signal arrival and return error immediately
	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sigChan:
			return fmt.Errorf("%w: %v", define.ErrCatchSignal, s)
		}
	})

	// Start a goroutine running api service,if catch error, return error
	g.Go(func() error {
		endPoint := filepath.Join(filepath.Dir(mc.Dirs.SocksDir.GetPath()), define.RESTAPIEndpointName)
		logrus.Infof("Start rest api service at %q", endPoint)
		return server.RestService(ctx, mc, endPoint) // server.RestService must now subscribe to ctx
	})

	// Start the machine, if catch error, return error
	// 1. start machine using shim.Start, the network provider(gvproxy) and Hypervisor(krunkit) as
	//    the child process of current process.
	// 2. if machine start successful, wait the network provider(gvproxy) ad Hypervisor(krunkit) exit
	//    this will block the current goroutine, until the network provider and Hypervisor exit
	// 3. if the network provider or Hypervisor got exit, clean up files and kill all child process
	g.Go(func() error {
		return startMachine(ctx, mc)
	})

	defer cleanUp(mc) // Clean tmp files at the end
	return g.Wait()   //nolint:wrapcheck
}

func startMachine(parentCtx context.Context, mc *vmconfig.MachineConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	context.AfterFunc(parentCtx, func() {
		SyncDisk(mc)
		cancel()
	})

	logrus.Infof("Starting machine %q", mc.VMName)
	// Start the network and hypervisor, shim.Start is non-block func
	_, err := shim.Start(ctx, mc)

	if err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}

	// Test the machine ssh connection by consume a string with '\n'
	if !shim.ConductVMReadinessCheck(ctx, mc) {
		return fmt.Errorf("machine ssh is not ready")
	}
	logrus.Infof("Machine %s SSH is ready, using sshkey %s with %s, listen in %d",
		mc.VMName, mc.SSH.IdentityPath, mc.SSH.RemoteUsername, mc.SSH.Port)

	// Test the podman api which forwarded from host to guest
	err = machine.WaitPodmanReady(ctx, mc.GvProxy.ForwardInfo["forward-sock"][0])
	if err != nil {
		return fmt.Errorf("failed to ping podman api: %w", err)
	}
	events.NotifyRun(events.Ready)
	logrus.Infof("Machine %s Podman API listened in %q", mc.VMName, mc.GvProxy.ForwardInfo["forward-sock"][0])

	// Start Sleep Notifier and dispatch tasks
	logrus.Infof("Start Sleep Notifier and dispatch tasks")
	go shim.SleepNotifier(ctx, mc)

	if err = mc.UpdateLastBoot(); err != nil {
		logrus.Warnf("failed to update last boot time: %v", err)
	}

	logrus.Infof("Machine start successful")

	return shim.Wait(ctx, mc) //nolint:wrapcheck
}

// cleanUp deletes the temporary socks file and terminates the child process using
// cmd.Process.Kill()
func cleanUp(mc *vmconfig.MachineConfig) {
	events.NotifyRun(events.SyncMachineDisk)

	logrus.Infof("Start clean up files")
	gvpBackendSocket, _ := mc.GVProxyNetworkBackendSocks()
	_ = gvpBackendSocket.Delete(true)

	gvpBackendSocket2 := &io.VMFile{Path: fmt.Sprintf("%s-%s", gvpBackendSocket.GetPath(), "krun.sock")}
	_ = gvpBackendSocket2.Delete(true)

	podmanInHost := mc.PodmanAPISocketHost()
	_ = podmanInHost.Delete(true)

	gvpPidFile := &io.VMFile{Path: mc.GvProxy.PidFile}
	_ = gvpPidFile.Delete(true)
}

func SyncDisk(mc *vmconfig.MachineConfig) {
	if err := shim.DiskSync(mc); err != nil {
		logrus.Warnf("Failed to sync disk: %v", err)
		return
	}
}
