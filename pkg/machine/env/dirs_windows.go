//go:build windows && (arm64 || amd64)

package env

import (
	"path/filepath"
)

// getTmpDir return ${BauklotzeHomePath}/tmp/
func getTmpDir() (string, error) {
	if CustomHomeEnv != "" {
		return filepath.Join(CustomHomeEnv, "tmp"), nil
	}
	p, err := GetBauklotzeHomePath()
	if err != nil {
		return "", err
	}

	// ${BauklotzeHomePath}/tmp/
	return filepath.Join(p, "tmp"), nil
}

// getRuntimeDir: return ${BauklotzeHomePath}/tmp/
func getRuntimeDir() (string, error) {
	return getTmpDir()
}
