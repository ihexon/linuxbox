//go:build windows && (arm64 || amd64)

package env

import (
	"os"
	"path/filepath"
)

func getTMPDir() (string, error) {
	if CustomHomeEnv != "" {
		return filepath.Join(CustomHomeEnv, ".tmp"), nil
	}

	tmpDir, ok := os.LookupEnv("TEMP")
	if !ok {
		tmpDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp")
	}
	return tmpDir, nil
}

func getRuntimeDir() (string, error) {
	return getTMPDir()
}
