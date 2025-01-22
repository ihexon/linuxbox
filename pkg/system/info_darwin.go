//go:build darwin

package system

import (
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

func Version() string {
	sysName, err := syscall.Sysctl("kern.ostype")
	if err != nil {
		logrus.Errorf("failed to get kernel os type: %v", err)
	}
	release, err := syscall.Sysctl("kern.osrelease")
	if err != nil {
		logrus.Errorf("failed to get kernel os release: %v", err)
	}
	version, err := syscall.Sysctl("kern.version")
	if err != nil {
		logrus.Errorf("failed to get kernel version: %v", err)
	}

	// The version might have newlines or tabs; convert to spaces.
	version = strings.ReplaceAll(version, "\n", " ")
	version = strings.ReplaceAll(version, "\t", " ")
	version = strings.TrimSpace(version)

	machine, err := syscall.Sysctl("hw.machine")
	if err != nil {
		logrus.Warnf("failed to get hardware machine: %v", err)
	}

	ret := sysName + " " + release + " " + version + " " + machine
	return ret
}
