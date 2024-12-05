//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"os"

	"github.com/containers/common/pkg/strongunits"
)

const (
	DefaultIdentityName  = "sshkey"
	MachineConfigVersion = 1
	// TODO: This should be configurable, in macos it should be a unix socket
)

type CreateVMOpts struct {
	Name          string       `json:"Name"`
	Dirs          *MachineDirs `json:"Dirs"`
	ReExec        bool         `json:"ReExec"`        // re-exec as administrator
	UserImageFile string       `json:"UserImageFile"` // Only used in wsl2
}

type WSLConfig struct {
}

type ResourceConfig struct {
	// CPUs to be assigned to the VM
	CPUs uint64 `json:"CPUs"`
	// Memory in megabytes assigned to the vm
	Memory strongunits.MiB `json:"Memory"`
}

type MachineDirs struct {
	ConfigDir     *VMFile `json:"ConfigDir"`
	DataDir       *VMFile `json:"DataDir"`
	ImageCacheDir *VMFile `json:"ImageCacheDir"`
	RuntimeDir    *VMFile `json:"RuntimeDir"`
	LogsDir       *VMFile `json:"LogsDir"`
}

const (
	DefaultMachineName = "bugbox-machine-default"
	DefaultUserInGuest = "root"
)

var (
	DefaultFilePerm os.FileMode = 0644
)

type StopOptions struct {
	SendEvt       string
	CommonOptions *CommonOptions
}

type InitOptions struct {
	IsDefault      bool
	CPUS           uint64
	Volumes        []string
	Memory         uint64
	Name           string
	Username       string
	ReExec         bool
	ImagesStruct   ImagesStruct
	ImageVerStruct ImageVerStruct
	CommonOptions  CommonOptions
}

type StartOptions struct {
	CommonOptions CommonOptions
}

type CommonOptions struct {
	ReportURL string
	PPID      int32
}

type ImageVerStruct struct {
	BootableImageVersion string
	DataDiskVersion      string
}

type ImagesStruct struct {
	BootableImage string // Bootable image
	DataDisk      string // Mounted in /var
}

type SetOptions struct {
	CPUs    uint64
	Memory  uint64
	Volumes []string
}

var (
	GitCommit string
)
