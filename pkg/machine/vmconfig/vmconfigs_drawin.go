//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package vmconfig

import (
	"bauklotze/pkg/machine/io"

	vfConfig "github.com/crc-org/vfkit/pkg/config"
	"github.com/sirupsen/logrus"
)

type Helper struct {
	LogLevel       logrus.Level             `json:"LogLevel"`
	Endpoint       string                   `json:"Endpoint"`
	BinaryPath     *io.VMFile               `json:"BinaryPath"`
	VirtualMachine *vfConfig.VirtualMachine `json:"VirtualMachine"`
}

type AppleKrunkitConfig struct {
	Krunkit Helper `json:"Krunkit"`
}

type AppleVFkitConfig struct {
	// The VFKit endpoint where we can interact with the VM
	Vfkit Helper `json:"Vfkit"`
}
