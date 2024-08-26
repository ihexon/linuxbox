package vmconfigs

type VolumeMountType int

const (
	// 9pfs
	NineP VolumeMountType = iota
	VirtIOFS
	Unknown
)
