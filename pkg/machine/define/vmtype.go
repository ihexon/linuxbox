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
	AppleHvVirt
	UnknownVirt
)
const (
	wsl     = "wsl"
	libkrun = "libkrun"
	appleHV = "applehv"
)

func (v VMType) String() string {
	switch v {
	case WSLVirt:
		return wsl
	case LibKrun:
		return libkrun
	case AppleHvVirt:
		return appleHV
	default:
	}
	return wsl
}

func ParseVMType(input string, emptyFallback VMType) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case wsl:
		return WSLVirt, nil
	case libkrun:
		return LibKrun, nil
	case appleHV:
		return AppleHvVirt, nil
	case "":
		return emptyFallback, nil
	default:
		return UnknownVirt, fmt.Errorf("unknown VMType `%s`", input)
	}
}
