//go:build darwin && !windows && !linux

package env

import "os"

// os.LookupEnv("TMPDIR") in macos return the path like `/var/folders/cc/5844tzj53ljcm_ph48hqlr8w0000gn/T/`,
// not /tmp !!
func getTMPDir() (string, error) {
	tmpDir, ok := os.LookupEnv("TMPDIR")
	if !ok {
		tmpDir = "/tmp"
	}
	return tmpDir, nil
}
