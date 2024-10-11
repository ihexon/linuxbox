//go:build darwin && (arm64 || amd64)

package ignition

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/vmconfigs"
	"fmt"
	"os"
	"strings"
)

func getLocalTimeZone() (string, error) {
	tzPath, err := os.Readlink("/etc/localtime")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(tzPath, "/var/db/timezone/zoneinfo"), nil
}

func (ign *DynamicIgnition) generateSetTimeZoneIgnitionCfg() error {
	tzInMacOS, err := getLocalTimeZone()
	if err != nil {
		return err
	}

	slice := config.NewSlice([]string{
		fmt.Sprintf("ln -sf /usr/share/zoneinfo/%s /etc/localtime", tzInMacOS),
		fmt.Sprintf("ln -sf /usr/share/zoneinfo/%s /usr/share/zoneinfo/localtime", tzInMacOS),
		"sync",
		"\n",
	})

	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}

func (ign *DynamicIgnition) generateMountsIgnitionCfg(mnts []*vmconfigs.Mount) error {
	var mountCmds []string
	for _, source := range mnts {
		mountCmds = append(mountCmds, fmt.Sprintf("mkdir -p %s", source.Target))
		mountCmds = append(mountCmds, fmt.Sprintf("mount -t virtiofs %s %s", source.Tag, source.Target))
		mountCmds = append(mountCmds, fmt.Sprint("\n"))
		mountCmds = append(mountCmds, fmt.Sprintf("%s", "sync"))
		mountCmds = append(mountCmds, fmt.Sprint("\n"))
	}

	slice := config.NewSlice(mountCmds)
	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}

// Write sshkey.pub into /root/.ssh/authorized_keys
func (ign *DynamicIgnition) generateSSHIgnitionCfg(sshkey_pub string) error {
	slice := config.NewSlice([]string{
		fmt.Sprintf("echo %s > /root/.ssh/authorized_keys", sshkey_pub),
		"\n",
		"sync",
		"\n",
	})
	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}

func (ign *DynamicIgnition) generateReadyEvent() error {
	slice := config.NewSlice([]string{
		fmt.Sprintf("%s", "/bin/echo Ready | /usr/bin/socat - VSOCK-CONNECT:2:1025"),
		"\n",
		"sync",
		"\n",
	})
	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}
