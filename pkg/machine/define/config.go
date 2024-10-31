package define

import (
	"github.com/containers/common/pkg/strongunits"
	"os"
)

const (
	DefaultIdentityName  = "sshkey"
	MachineConfigVersion = 1
	// TODO: This should be configurable, in macos it should be a unix socket
	TcpIgnitionAddr = "tcp://127.0.0.1:65530"
	CliProxyUDF     = "/tmp/cli_proxy.sock"
)

type CreateVMOpts struct {
	Name          string
	Dirs          *MachineDirs
	ReExec        bool   // re-exec as administrator
	UserImageFile string // Only used in wsl2
}

type WSLConfig struct {
}

type ResourceConfig struct {
	// CPUs to be assigned to the VM
	CPUs uint64
	// Memory in megabytes assigned to the vm
	Memory strongunits.MiB
}

type MachineDirs struct {
	ConfigDir     *VMFile
	DataDir       *VMFile
	ImageCacheDir *VMFile
	RuntimeDir    *VMFile
	LogsDir       *VMFile
}

const (
	DefaultMachineName string = "bugbox-machine-default"
	DefaultUserInGuest        = "root"
)

var (
	DefaultFilePerm os.FileMode = 0644
)

type StopOptions struct {
	SendEvt       string
	CommonOptions *CommonOptions
}

type InitOptions struct {
	IsDefault     bool
	CPUS          uint64
	Volumes       []string
	Memory        uint64
	Name          string
	Username      string
	ReExec        bool
	Images        ImagesStruct
	ImageVersion  ImageVerStruct
	PPID          int32
	CommonOptions CommonOptions
}

type StartOptions struct {
	CommonOptions CommonOptions
}

type CommonOptions struct {
	ReportUrl string
	PPID      int32
}

type ImageVerStruct struct {
	BootableImageVersion string
	DataDiskVersion      string
}

type ImagesStruct struct {
	BootableImage string // Bootable image
	DataDisk      string // Mounted in /var
	OverlayImage  string // Overlay image mounted /
}

type SetOptions struct {
	CPUs    uint64
	Memory  uint64
	Volumes []string
}
