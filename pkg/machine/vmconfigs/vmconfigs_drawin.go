package vmconfigs

import "bauklotze/pkg/machine/apple/vfkit"

type HyperVConfig struct{}
type WSLConfig struct{}
type QEMUConfig struct{}

type AppleHVConfig struct {
	// The VFKit endpoint where we can interact with the VM
	Vfkit vfkit.Helper
}
