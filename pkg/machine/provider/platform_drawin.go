package provider

import "bauklotze/pkg/machine/vmconfigs"

func GetAll(_ bool) ([]vmconfigs.VMProvider, error) {
	return []vmconfigs.VMProvider{
		new(applehv.AppleHVStubber),
		new(libkrun.LibKrunStubber),
	}, nil
}
