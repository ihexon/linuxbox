//go:build darwin && (arm64 || amd64)

package shim

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

const (
	defaultGuestSock = "/run/podman/podman.sock"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *define.MachineDirs) ([]string, string, machine.APIForwardingState, error) {
	// all hostSocketIO will be forward in to Guest forwardSock
	hostSocket, err := mc.APISocket()
	if err != nil {
		return nil, "", machine.NoForwarding, err
	}

	forwardSock, state, err := setupForwardingLinks(hostSocket, dirs.DataDir)
	if err != nil {
		return nil, "", machine.NoForwarding, err
	}
	return []string{hostSocket.GetPath()}, forwardSock, state, nil
}

func setupForwardingLinks(hostSocket, dataDir *define.VMFile) (string, machine.APIForwardingState, error) {
	_ = hostSocket.Delete()
	return hostSocket.GetPath(), machine.NotInstalled, nil
}

// Note that mc is a **Point** to the vmconfigs.MachineConfig
func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *define.MachineDirs, hostSocks []string) error {
	forwardUser := mc.SSH.RemoteUsername

	guestSock := defaultGuestSock

	cfg := config.Default()

	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName, false)
	if err != nil {
		return err
	}
	cmd := gvproxy.NewGvproxyCommand() // New a GvProxyCommands
	runDir := dirs.RuntimeDir

	cmd.PidFile = filepath.Join(runDir.GetPath(), "gvproxy.pid")
	cmd.LogFile = filepath.Join(runDir.GetPath(), "gvproxy.log")

	cmd.SSHPort = mc.SSH.Port

	// For now we only have one hostSocks that is podman api
	for _, hostSock := range hostSocks {
		cmd.AddForwardSock(hostSock)
		cmd.AddForwardDest(guestSock)
		cmd.AddForwardUser(forwardUser)
		cmd.AddForwardIdentity(mc.SSH.IdentityPath)
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		cmd.Debug = true
		logrus.Debug(cmd)
	}

	if err := provider.StartNetworking(mc, &cmd); err != nil {
		return err
	}
	gvcmd := cmd.Cmd(binary)

	logrus.Infof("Gvproxy command-line: %s %s", binary, strings.Join(cmd.ToCmdline(), " "))
	if err := gvcmd.Start(); err != nil {
		return fmt.Errorf("unable to execute: %q: %w", cmd.ToCmdline(), err)
	}

	machine.GlobalPIDs.SetGvproxyPID(gvcmd.Process.Pid)
	machine.GlobalCmds.SetGvpCmd(gvcmd)

	mc.GVProxyPid = int32(gvcmd.Process.Pid)
	mc.GvProxy.GvProxy.PidFile = cmd.PidFile
	mc.GvProxy.GvProxy.LogFile = cmd.LogFile
	mc.GvProxy.GvProxy.SSHPort = cmd.SSHPort
	mc.GvProxy.GvProxy.MTU = cmd.MTU
	mc.GvProxy.HostSocks = hostSocks
	mc.GvProxy.RemoteSocks = guestSock

	return nil
}
