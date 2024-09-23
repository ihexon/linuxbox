//go:build darwin && (arm64 || amd64)

package ignition

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine"
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
	})

	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}

// TODO: generateMountsIgnitionCfg
func (ign *DynamicIgnition) generateMountsIgnitionCfg(mounts []machine.VirtIoFs) error {
	slice := config.NewSlice([]string{})
	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}

// Write sshkey.pub into /root/.ssh/authorized_keys
func (ign *DynamicIgnition) generateSSHIgnitionCfg(sshkey_pub string) error {
	slice := config.NewSlice([]string{
		fmt.Sprintf("echo %s > /root/.ssh/authorized_keys", sshkey_pub),
		"sync",
	})
	ign.Commands = append(ign.Commands, slice.Get()...)
	return nil
}
