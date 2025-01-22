//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package defconfig

import (
	"fmt"
	"strings"
)

type VMType int64

const (
	LibKrun VMType = iota
	VFkit
	UnknownVirt
)

const (
	libkrun = "libkrun"
	vfkit   = "vfkit"
)

func (v VMType) String() string {
	switch v {
	case LibKrun:
		return libkrun
	case VFkit:
		return vfkit
	default:
	}
	return ""
}

// ParseVMType covert string to VMType (int64)
func ParseVMType(input string) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case libkrun:
		return LibKrun, nil
	case vfkit:
		return VFkit, nil
	default:
		return UnknownVirt, fmt.Errorf("unknown VMType `%s`", input)
	}
}
