package vmconfig

import (
	"fmt"
	"strings"
)

type VMType int64

const (
	LibKrun VMType = iota
	VFkit
	Unknown
)

const (
	libkrun = "libkrun"
	krunkit = "krunkit"
	vfkit   = "vfkit"
	unknown = "unknown"
)

func (v VMType) String() string {
	switch v {
	case LibKrun:
		return libkrun
	case VFkit:
		return vfkit
	default:
		return unknown
	}
}

// ParseVMType converts a string to VMType (int64)
func ParseVMType(input string) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case libkrun, krunkit:
		return LibKrun, nil
	case vfkit:
		return VFkit, nil
	default:
		return Unknown, fmt.Errorf("unknown VMType %q", input)
	}
}
