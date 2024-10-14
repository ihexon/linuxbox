package define

import (
	strongunits "github.com/containers/common/pkg/strongunits"
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
	ReExec        bool // re-exec as administrator
	UserImageFile string
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
}

type ResetOptions struct {
	Force bool
}

const (
	DefaultMachineName string = "bugbox-machine-default"
	DefaultUserInGuest        = "root"
)

var (
	DefaultFilePerm os.FileMode = 0644
)

type StartOptions struct {
	TwinPid   int32
	SendEvt   string
	Volumes   []string
	ReportUrl string
}

type StopOptions struct {
	SendEvt string
}

type InitOptions struct {
	IsDefault    bool
	CPUS         uint64
	Volumes      []string
	Memory       uint64
	Name         string
	Username     string
	ReExec       bool
	SendEvt      string
	Images       ImagesStruct
	ImageVersion ImageVerStruct
}

type ImageVerStruct struct {
	BootableImageVersion string
	ExternalDiskVersion  string
}

type ImagesStruct struct {
	BootableImage string
	ExternalDisk  string
}

type SetOptions struct {
	CPUs    uint64
	Memory  uint64
	Volumes []string
}
