//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

//go:build darwin

package config

import (
	"runtime"
	"sync"

	"bauklotze/pkg/machine/define"
)

// Options to use when loading a Config via New().
type Options struct {
	SetDefault bool
}

var (
	errCachedConfig   error
	cachedConfigMutex sync.Mutex
	cachedConfig      *Config
)

func New(options *Options) *Config {
	if options == nil {
		options = &Options{}
	} else if options.SetDefault {
		cachedConfigMutex.Lock()
		defer cachedConfigMutex.Unlock()
	}
	return newLocked(options)
}

func newLocked(options *Options) *Config {
	// Start with the built-in defaults
	config := defaultConfig()

	if options.SetDefault {
		cachedConfig = config
		errCachedConfig = nil
	}
	return config
}

func Default() *Config {
	cachedConfigMutex.Lock()
	defer cachedConfigMutex.Unlock()
	if cachedConfig != nil || errCachedConfig != nil {
		return cachedConfig
	}
	cachedConfig = newLocked(&Options{SetDefault: true})
	return cachedConfig
}

func getDefaultMachineUser() string {
	return define.DefaultUserInGuest
}

const (
	defaultDiskSize     = 100
	defaultMemory       = 2048
	defaultDataDiskSize = 100
)

// defaultMachineConfig returns the default machine configuration.
func defaultMachineConfig() MachineConfig {
	cpus := runtime.NumCPU() / 2 //nolint:mnd
	if cpus == 0 {
		cpus = 1
	}
	return MachineConfig{
		CPUs:         uint64(cpus),
		DiskSize:     defaultDiskSize,
		Image:        "",
		Memory:       defaultMemory,
		DataDiskSize: defaultDataDiskSize,
		Volumes:      NewSlice(getDefaultMachineVolumes()),
		User:         getDefaultMachineUser(), // I tell u a joke :)
	}
}
