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
