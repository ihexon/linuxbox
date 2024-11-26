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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	podmanGuestSocks = "/run/podman/podman.sock"
)

func setupMachineSockets(mc *vmconfigs.MachineConfig, dirs *define.MachineDirs) (string, string, error) {
	podmanApiSocketOnHost, err := mc.PodmanApiSocketHost()
	if err != nil {
		return "", "", err
	}
	err = podmanApiSocketOnHost.Delete()
	if err != nil {
		return "", "", err
	}
	return podmanApiSocketOnHost.GetPath(), podmanGuestSocks, nil
}

func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *define.MachineDirs, socksInHost string, socksInGuest string) (*exec.Cmd, error) {
	forwardUser := mc.SSH.RemoteUsername

	cfg := config.Default()

	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName)
	if err != nil {
		return nil, err
	}

	cmd := gvproxy.NewGvproxyCommand() // New a GvProxyCommands
	runDir := dirs.RuntimeDir
	logsDIr := dirs.LogsDir

	cmd.PidFile = filepath.Join(runDir.GetPath(), "gvproxy.pid")
	cmd.LogFile = filepath.Join(logsDIr.GetPath(), "gvproxy.log")

	cmd.SSHPort = mc.SSH.Port
	cmd.AddForwardSock(socksInHost)             // podman api in host
	cmd.AddForwardDest(socksInGuest)            // podman api in guest
	cmd.AddForwardUser(forwardUser)             // always be root
	cmd.AddForwardIdentity(mc.SSH.IdentityPath) // ssh keys

	if err := provider.StartNetworking(mc, &cmd); err != nil {
		return nil, err
	}

	gvcmd := cmd.Cmd(binary)
	gvcmd.Stdout = os.Stdout
	gvcmd.Stderr = os.Stderr

	logrus.Infof("Gvproxy command-line: %s %s", binary, strings.Join(cmd.ToCmdline(), " "))
	if err := gvcmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to execute: %q: %w", cmd.ToCmdline(), err)
	} else {
		machine.GlobalCmds.SetGvpCmd(gvcmd)
	}

	mc.GvProxy.GvProxy.PidFile = cmd.PidFile
	mc.GvProxy.GvProxy.LogFile = cmd.LogFile
	mc.GvProxy.GvProxy.SSHPort = cmd.SSHPort
	mc.GvProxy.GvProxy.MTU = cmd.MTU
	mc.GvProxy.HostSocks = []string{socksInHost}
	mc.GvProxy.RemoteSocks = socksInGuest

	return gvcmd, mc.Write()
}
