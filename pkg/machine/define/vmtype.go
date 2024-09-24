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
	UnknownVirt
)
const (
	wsl     = "wsl"
	libkrun = "libkrun"
)

func (v VMType) String() string {
	switch v {
	case WSLVirt:
		return wsl
	case LibKrun:
		return libkrun
	default:
	}
	return wsl
}

type ImageFormat int64

const (
	Vhdx ImageFormat = iota
	Tar
	Raw
)

func (imf ImageFormat) Kind() string {
	switch imf {
	case Vhdx:
		return "vhdx"
	case Tar:
		return "tar" //  wsl2_rootfs.tar
	case Raw:
		return "raw"
	}
	return "raw"
}

func (imf ImageFormat) KindWithCompression() string {
	// Tar uses xz; all others use zstd
	if imf == Tar {
		return "tar.xz" // wsl2_rootfs.tar.xz
	}
	return fmt.Sprintf("%s.zst", imf.Kind()) // wsl2_rootfs.tar.zst
}

func (v VMType) ImageFormat() ImageFormat {
	switch v {
	case WSLVirt:
		return Tar
	case LibKrun:
		return Raw
	}
	return Raw
}

func ParseVMType(input string, emptyFallback VMType) (VMType, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case wsl:
		return WSLVirt, nil
	case libkrun:
		return LibKrun, nil
	case "":
		return emptyFallback, nil
	default:
		return UnknownVirt, fmt.Errorf("unknown VMType `%s`", input)
	}
}
