//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"fmt"
	"strings"
)

// VMType OK for now
type VMType int64

const (
	WSLVirt VMType = iota
	LibKrun
	VFkit
	UnknownVirt
)
const (
	wsl     = "wsl"
	libkrun = "libkrun"
	vfkit   = "vfkit"
)

func (v VMType) String() string {
	switch v {
	case WSLVirt:
		return wsl
	case LibKrun:
		return libkrun
	case VFkit:
		return vfkit
	default:
	}
	return wsl
}

func ParseVMType(input string, fallback VMType) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case wsl:
		return WSLVirt, nil
	case libkrun:
		return LibKrun, nil
	case vfkit:
		return VFkit, nil
	case "":
		return fallback, nil
	default:
		return UnknownVirt, fmt.Errorf("unknown VMType `%s`", input)
	}
}
