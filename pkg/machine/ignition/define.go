//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

const MountOpenrcTemplate = `#!/sbin/openrc-run

start() {
	ebegin "Mounting {{.Source}} to {{.Target}}"
	mkdir -p "{{.Target}}"
	mount -t "{{.FsType}}" "{{.Source}}" "{{.Target}}" || return 1
	eend $?
}

stop() {
	ebegin "Unmounting {{.Target}}"
	umount "{{.Target}}" || return 1
	eend $?
}
`

const (
	virtiofs = "virtiofs"
	rcPrefix = "ovm_"
)
