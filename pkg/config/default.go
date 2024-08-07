package config

import (
	"runtime"
)

func defaultConfig() (*Config, error) {
	return &Config{
		Machine: defaultMachineConfig(),
	}, nil
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
		Volumes:  Slice{Values: getEmptyMachineVolumes()},
		User:     getDefaultMachineUser(), // I tell u a joke :)
	}
}

func getDefaultMachineUser() string {
	return "donaldtrump"
}
