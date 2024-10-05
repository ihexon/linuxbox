//go:build darwin && !windows && !linux

package env

import "path/filepath"

// getTmpDir return ${BauklotzeHomePath}/tmp/
func getTmpDir() (string, error) {
	if CustomHomeEnv != "" {
		return filepath.Join(CustomHomeEnv, "tmp"), nil
	}
	p, err := GetBauklotzeHomePath()
	if err != nil {
		return "", err
	}

	return filepath.Join(p, "tmp"), nil // ${BauklotzeHomePath}/tmp/
}

// getRuntimeDir: ${BauklotzeHomePath}/tmp/
func getRuntimeDir() (string, error) {
	return getTmpDir() // ${BauklotzeHomePath}/tmp/
}
