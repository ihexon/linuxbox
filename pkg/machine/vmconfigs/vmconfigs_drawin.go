//go:build (darwin || linux) && (amd64 || arm64)

package vmconfigs

import "bauklotze/pkg/machine/apple/hvhelper"

type HyperVConfig struct{}
type WSLConfig struct{}
type QEMUConfig struct{}

// krunkit 的优先级放到最高
type AppleKrunkitConfig struct {
	Krunkit hvhelper.Helper
}
