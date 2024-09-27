package shim

import (
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/vmconfigs"
)

func CmdLineVolumesToMounts(volumes []string, volumeType vmconfigs.VolumeMountType) []*vmconfigs.Mount {
	mounts := []*vmconfigs.Mount{}
	for i, volume := range volumes {
		if volume == "" {
			continue
		}
		var mount vmconfigs.Mount
		tag, source, target, readOnly, _ := vmconfigs.SplitVolume(i, volume)
		switch volumeType {
		case vmconfigs.VirtIOFS:
			mount = machine.NewVirtIoFsMount(source, target, readOnly).ToMount()
		default:
			mount = vmconfigs.Mount{
				Type:          volumeType.String(),
				Tag:           tag,
				Source:        source,
				Target:        target,
				ReadOnly:      readOnly,
				OriginalInput: volume,
			}
		}
		mounts = append(mounts, &mount)
	}
	return mounts
}
