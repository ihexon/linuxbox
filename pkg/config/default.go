package config

import (
	"runtime"
)

func Default() (*Config, error) {
	config, err := defaultConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

// defaultConfig return a &Config
func defaultConfig() (*Config, error) {
	return &Config{
		Machine: defaultMachineConfig(),
	}, nil
}

func getDefaultMachineUser() string {
	return "donaldtrump"
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
