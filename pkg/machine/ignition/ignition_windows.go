//go:build windows && (arm64 || amd64)

package ignition

import (
	"bauklotze/pkg/machine/vmconfigs"
)

func getLocalTimeZone() (string, error) {
	return "", nil
}
func (ign *DynamicIgnition) generateSetTimeZoneIgnitionCfg() error {
	return nil
}

func (ign *DynamicIgnition) generateMountsIgnitionCfg(mnts []*vmconfigs.Mount) error {

	return nil
}

// Write sshkey.pub into /root/.ssh/authorized_keys
func (ign *DynamicIgnition) generateSSHIgnitionCfg(sshkey_pub string) error {
	return nil
}

func (ign *DynamicIgnition) generateReadyEvent() error {

	return nil
}
