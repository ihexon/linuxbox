//go:build darwin && !windows && !linux

package env

import (
	"os"
)

// os.LookupEnv("TMPDIR") in macos return the path like `/var/folders/cc/5844tzj53ljcm_ph48hqlr8w0000gn/T/`,
// if CustomHomeDir is set, then return CustomHomeDIr/.tmp
func getTMPDir() (string, error) {
	if CustomHomeEnv != "" {
		return CustomHomeEnv + ".tmp", nil
	}

	tmpDir, ok := os.LookupEnv("TMPDIR")
	if !ok {
		tmpDir = "/tmp"
	}
	return tmpDir, nil
}

func getRuntimeDir() (string, error) {
	return getTMPDir()
}
