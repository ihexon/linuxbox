package define

import (
	lockfile "bauklotze/pkg/filelock"
	"bauklotze/pkg/ioutils"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const MachineConfigVersion = 1

type CreateVMOpts struct {
	Name               string
	Dirs               *MachineDirs
	ReExec             bool // re-exec as administrator
	UserModeNetworking bool
}

type GiB uint64
type MiB uint64
type cores uint64

type WSLConfig struct {
	// Uses usermode networking
	UserModeNetworking bool
}

type ResourceConfig struct {
	// CPUs to be assigned to the VM
	CPUs uint64
	// Disk size in gigabytes assigned to the vm
	DiskSize GiB
	// Memory in megabytes assigned to the vm
	Memory MiB
}

type machineImage interface { //nolint:unused
	download() error
	path() string
}

type MachineDirs struct {
	ConfigDir     *VMFile
	DataDir       *VMFile
	ImageCacheDir *VMFile
	RuntimeDir    *VMFile
}

type VMFile struct {
	Path string
}

type ResetOptions struct {
	Force bool
}

const (
	DefaultMachineName string = "bugbox-machine-default"
)

// MachineConfig This is a more detailed, provider-specific configuration structure. \
// It includes specific settings for different virtualization providers like QEMU, HyperV, AppleHV, and WSL.
// It contains fields like QEMUHypervisor, HyperVHypervisor, etc., which are tailored to each provider's requirements.
//
// It allows Podman to handle the unique requirements and features of each virtualization technology.
// MachineConfig is used when dealing with provider-specific operations and configurations.
type MachineConfig struct {
	Created          time.Time
	Dirs             *MachineDirs
	Name             string
	ImagePath        *VMFile    // 实际上是 rootfs 的路径
	WSLHypervisor    *WSLConfig `json:",omitempty"`
	Starting         bool
	ConfigPath       *VMFile
	Resources        ResourceConfig
	imageDescription machineImage
	Version          uint
	Lock             *lockfile.LockFile
}

var (
	DefaultFilePerm os.FileMode = 0644
)

func (mc *MachineConfig) Write() error {
	if mc.ConfigPath == nil {
		return fmt.Errorf("no configuration file associated with vm %q", mc.Name)
	}
	b, err := json.Marshal(mc)
	if err != nil {
		return err
	}
	logrus.Debugf("writing configuration file %q", mc.ConfigPath.Path)
	return ioutils.AtomicWriteFile(mc.ConfigPath.GetPath(), b, DefaultFilePerm)
}

type InitOptions struct {
	CPUS     uint64
	DiskSize uint64
	Image    string
	Volumes  []string
	Memory   uint64
	Name     string
	Username string
	ReExec   bool
}

type VMProvider interface { //nolint:interfacebloat
	VMType() VMType
	Exists(name string) (bool, error)
	GetDisk(userInputPath string, dirs *MachineDirs, mc *MachineConfig) error
	CreateVM(opts CreateVMOpts, mc *MachineConfig) error
	StopVM(mc *MachineConfig, hardStop bool) error
}
