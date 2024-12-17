//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

const VirtioFSMountScript = `
echo "Mounting {{.Source}} Tag {{.Tag}} to {{.Target}}"
mkdir -p "{{.Target}}"
mount -t "{{.FsType}}" "{{.Tag}}" "{{.Target}}" || echo "Error: Mounting {{.Source}} to {{.Target}} failed"
`

const WriteSSHPubKeyScript = `
echo "Writing SSH public key to /root/.ssh/authorized_keys"
mkdir -p "/root/.ssh/"
echo "{{.Target}}" >> "/root/.ssh/authorized_keys"
`

const UpdateTimeZoneScript = `
echo "Setting timezone to {{.TimeZone}}"
ln -sf "/usr/share/zoneinfo/{{.TimeZone}}" "/etc/localtime"
`

const podmanMachineConfigScript = `
echo "Generating podman machine config"
echo "{{.CurrentVMType}}" > "/etc/containers/podman-machine"
`
