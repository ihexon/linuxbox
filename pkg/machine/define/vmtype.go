package define

import (
	"fmt"
	"strings"
)

// VMType OK for now
type VMType int64

const (
	QemuVirt VMType = iota
	WSLVirt
	AppleHvVirt
	HyperVVirt
	LibKrun
	UnknownVirt
)
const (
	wsl     = "wsl"
	qemu    = "qemu"
	appleHV = "applehv"
	hyperV  = "hyperv"
	libkrun = "libkrun"
)

func (v VMType) String() string {
	switch v {
	case WSLVirt:
		return wsl
	case AppleHvVirt:
		return appleHV
	case HyperVVirt:
		return hyperV
	case LibKrun:
		return libkrun
	}
	return qemu
}

func ParseVMType(input string, emptyFallback VMType) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case qemu:
		return QemuVirt, nil
	case wsl:
		return WSLVirt, nil
	case appleHV:
		return AppleHvVirt, nil
	case libkrun:
		return LibKrun, nil
	case hyperV:
		return HyperVVirt, nil
	case "":
		return emptyFallback, nil
	default:
		return UnknownVirt, fmt.Errorf("unknown VMType `%s`", input)
	}
}
