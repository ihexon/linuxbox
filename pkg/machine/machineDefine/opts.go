package machineDefine

import strongunits "bauklotze/pkg/storage"

type SetOptions struct {
	CPUs               *uint64
	DiskSize           *strongunits.GiB
	Memory             *strongunits.MiB
	Rootful            *bool
	UserModeNetworking *bool
	USBs               *[]string
}
