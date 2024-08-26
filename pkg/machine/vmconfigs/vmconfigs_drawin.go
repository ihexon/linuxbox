package vmconfigs

import "bauklotze/pkg/machine/apple/vfkit"

type HyperVConfig struct{}
type WSLConfig struct{}
type QEMUConfig struct{}

// AppleVfkitConfig describes the use of vfkit: cmdline and endpoint
type AppleVfkitConfig struct {
	// The VFKit endpoint where we can interact with the VM
	//  vfkit.Helper describes the use of vfkit: cmdline and endpoint
	Vfkit vfkit.Helper
}
