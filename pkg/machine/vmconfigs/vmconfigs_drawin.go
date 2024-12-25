//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vmconfigs

import (
	"bauklotze/pkg/machine/apple/hvhelper"
)

type AppleKrunkitConfig struct {
	Krunkit hvhelper.Helper `json:"Krunkit"`
}

type AppleVFkitConfig struct {
	// The VFKit endpoint where we can interact with the VM
	Vfkit hvhelper.Helper `json:"Vfkit"`
}
