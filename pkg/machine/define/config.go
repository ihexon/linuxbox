package define

import (
	"github.com/containers/common/pkg/strongunits"
	"os"
)

const (
	DefaultIdentityName  = "sshkey"
	MachineConfigVersion = 1
	MyName               = "Bauklotze"
	WorkDir              = "." + MyName + "_dir"
	TcpIgnitionAddr      = "tcp://127.0.0.1:65530"
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

type StartOptions struct {
	PPID          int32
	CommonOptions CommonOptions
}

type StopOptions struct {
	SendEvt       string
	CommonOptions CommonOptions
}

type InitOptions struct {
	PPID          int32
	IsDefault     bool
	CPUS          uint64
	Volumes       []string
	Memory        uint64
	Name          string
	Username      string
	ReExec        bool
	Images        ImagesStruct
	ImageVersion  ImageVerStruct
	CommonOptions CommonOptions
}

type CommonOptions struct {
	ReportUrl string
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
