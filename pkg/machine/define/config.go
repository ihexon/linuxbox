package define

import (
	strongunits "bauklotze/pkg/storage"
	"os"
)

const (
	DefaultIdentityName  = "sshkey"
	MachineConfigVersion = 1
	MyName               = "Bauklotze"
)

type CreateVMOpts struct {
	Name          string
	Dirs          *MachineDirs
	ReExec        bool // re-exec as administrator
	UserImageFile string
}

type GiB uint64
type MiB uint64
type cores uint64

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
)

var (
	DefaultFilePerm os.FileMode = 0644
)

type StartOptions struct {
	WaitAndStop bool // NoQuit when machine start
	TwinPid     int
	SendEvt     string
}

type StopOptions struct {
	TwinPid int
	SendEvt string
}

type InitOptions struct {
	IsDefault    bool
	CPUS         uint64
	Image        string
	Volumes      []string
	Memory       uint64
	Name         string
	Username     string
	ReExec       bool
	TwinPid      int
	SendEvt      string
	ImageVersion string
}

type SetOptions struct {
	CPUs    uint64
	Memory  uint64
	Volumes []string
}
