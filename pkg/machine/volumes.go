package machine

import "bauklotze/pkg/machine/vmconfigs"

type VolumeKind string

var (
	VirtIOFsVk VolumeKind = "virtiofs"
	NinePVk    VolumeKind = "9p"
)

type VirtIoFs struct {
	VolumeKind
	ReadOnly bool
	Source   string
	Tag      string
	Target   string
}

func MountToVirtIOFs(mnt *vmconfigs.Mount) VirtIoFs {
	return VirtIoFs{
		VolumeKind: VirtIOFsVk,
		ReadOnly:   mnt.ReadOnly,
		Source:     mnt.Source,
		Tag:        mnt.Tag,
		Target:     mnt.Target,
	}
}
