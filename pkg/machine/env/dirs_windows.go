//go:build windows && (arm64 || amd64)

package env

import "os"

func getTMPDir() (string, error) {
	tmpDir, ok := os.LookupEnv("TEMP")
	if !ok {
		tmpDir = os.Getenv("LOCALAPPDATA") + "\\Temp"
	}
	return tmpDir, nil
}

func getRuntimeDir() (string, error) {
	return getTMPDir()
}
