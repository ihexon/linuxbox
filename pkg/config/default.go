package config

import (
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

func New(options *Options) (*Config, error) {
	if options == nil {
		options = &Options{}
	} else if options.SetDefault {
		cachedConfigMutex.Lock()
		defer cachedConfigMutex.Unlock()
	}
	return newLocked(options)
}

func newLocked(options *Options) (*Config, error) {
	// Start with the built-in defaults
	config, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	if options.SetDefault {
		cachedConfig = config
		cachedConfigError = nil
	}

	return config, nil
}

func Default() (*Config, error) {
	cachedConfigMutex.Lock()
	defer cachedConfigMutex.Unlock()
	if cachedConfig != nil || cachedConfigError != nil {
		return cachedConfig, cachedConfigError
	}
	cachedConfig, cachedConfigError = newLocked(&Options{SetDefault: true})
	return cachedConfig, cachedConfigError
}

func defaultConfig() (*Config, error) {
	c := &Config{Machine: defaultMachineConfig()}
	c.Machine.HelperBinariesDir.Set(defaultHelperBinariesDir)
	return c, nil
}

func getDefaultMachineUser() string {
	return "root"
}

// defaultMachineConfig returns the default machine configuration.
func defaultMachineConfig() MachineConfig {
	cpus := runtime.NumCPU() / 2
	if cpus == 0 {
		cpus = 1
	}
	return MachineConfig{
		CPUs:     uint64(cpus),
		DiskSize: 100,
		Image:    "",
		Memory:   2048,
		Volumes:  NewSlice(getDefaultMachineVolumes()),
		User:     getDefaultMachineUser(), // I tell u a joke :)
	}
}
