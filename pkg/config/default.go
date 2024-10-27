package config

import (
	"bauklotze/pkg/machine/define"
	"runtime"
	"sync"
)

// Options to use when loading a Config via New().
type Options struct {
	SetDefault bool
}

var (
	cachedConfigError error
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
		cachedConfigError = nil
	}
	return config
}

func Default() *Config {
	cachedConfigMutex.Lock()
	defer cachedConfigMutex.Unlock()
	if cachedConfig != nil || cachedConfigError != nil {
		return cachedConfig
	}
	cachedConfig = newLocked(&Options{SetDefault: true})
	return cachedConfig
}

func getDefaultMachineUser() string {
	return define.DefaultUserInGuest
}

// defaultMachineConfig returns the default machine configuration.
func defaultMachineConfig() MachineConfig {
	cpus := runtime.NumCPU() / 2
	if cpus == 0 {
		cpus = 1
	}
	return MachineConfig{
		CPUs:         uint64(cpus),
		DiskSize:     100,
		Image:        "",
		Memory:       2048,
		DataDiskSize: 100,
		Volumes:      NewSlice(getDefaultMachineVolumes()),
		User:         getDefaultMachineUser(), // I tell u a joke :)
	}
}
