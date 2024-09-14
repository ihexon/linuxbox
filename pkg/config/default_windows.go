//go:build windows

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getDefaultMachineVolumes() []string {
	hd, _ := os.UserHomeDir()
	vol := filepath.VolumeName(hd)
	hostMnt := filepath.ToSlash(strings.TrimPrefix(hd, vol))
	return []string{fmt.Sprintf("%s:%s", hd, hostMnt)}
}

var defaultHelperBinariesDir = []string{
	"C:\\Program Files\\RedHat\\Podman",
}
