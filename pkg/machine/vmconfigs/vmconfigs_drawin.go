//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build (darwin || linux) && (amd64 || arm64)

package vmconfigs

import "bauklotze/pkg/machine/apple/hvhelper"

type HyperVConfig struct{}
type WSLConfig struct{}
type QEMUConfig struct{}

// krunkit 的优先级放到最高
type AppleKrunkitConfig struct {
	Krunkit hvhelper.Helper `json:"Krunkit"`
}
