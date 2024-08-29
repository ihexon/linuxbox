package vmconfigs

import "bauklotze/pkg/machine/apple/hvhelper"

type HyperVConfig struct{}
type WSLConfig struct{}
type QEMUConfig struct{}

// AppleVfkitConfig describes the use of hvhelper: cmdline and endpoint
// 暂停开发 hvhelper 的 provider
type AppleVfkitConfig struct {
	Vfkit hvhelper.Helper
}

// krunkit 的优先级放到最高
type AppleKrunkitConfig struct {
	Krunkit hvhelper.Helper
}
