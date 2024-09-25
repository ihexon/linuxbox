//go:build darwin && arm64

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

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *define.MachineDirs) ([]string, string, machine.APIForwardingState, error) {
	hostSocket, err := mc.APISocket()
	if err != nil {
		return nil, "", 0, err
	}

	forwardSock, state, err := setupForwardingLinks(hostSocket, dirs.DataDir)
	if err != nil {
		return nil, "", 0, err
	}
	return []string{hostSocket.GetPath()}, forwardSock, state, nil
}

func setupForwardingLinks(hostSocket, dataDir *define.VMFile) (string, machine.APIForwardingState, error) {
	_ = hostSocket.Delete()
	return hostSocket.GetPath(), machine.NotInstalled, nil
}

func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *define.MachineDirs, hostSocks []string) error {
	forwardUser := mc.SSH.RemoteUsername

	// TODO should this go up the stack higher or
	// the guestSock is "inside" the guest machine
	guestSock := fmt.Sprintf(defaultGuestSock, mc.HostUser.UID)
	if mc.HostUser.Rootful {
		guestSock = "/run/podman/podman.sock"
		forwardUser = "root"
	}

	cfg, err := config.Default()
	if err != nil {
		return err
	}

	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName, false)
	if err != nil {
		return err
	}
	cmd := gvproxy.NewGvproxyCommand()
	runDir := dirs.RuntimeDir
	cmd.PidFile = filepath.Join(runDir.GetPath(), "gvproxy.pid")
	cmd.LogFile = filepath.Join(runDir.GetPath(), "gvproxy.log")
	cmd.SSHPort = mc.SSH.Port

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
	c := cmd.Cmd(binary)

	logrus.Infof("gvproxy command-line: %s %s", binary, strings.Join(cmd.ToCmdline(), " "))
	if err := c.Start(); err != nil {
		return fmt.Errorf("unable to execute: %q: %w", cmd.ToCmdline(), err)
	}
	machine.GlobalPIDs.SetGvproxyPID(c.Process.Pid)
	return nil
}
