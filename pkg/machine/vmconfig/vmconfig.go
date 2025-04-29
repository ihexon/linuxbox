//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package vmconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bauklotze/pkg/machine/define"
	io2 "bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/volumes"
	"bauklotze/pkg/machine/workspace"
	"bauklotze/pkg/port"

	"github.com/containers/storage/pkg/ioutils"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type VMState struct {
	SSHReady    bool
	PodmanReady bool
}

type VMProvider interface { //nolint:interfacebloat
	InitializeVM(opts VMOpts) (*MachineConfig, error)
	StartNetworkProvider(ctx context.Context, mc *MachineConfig) error
	StartVMProvider(ctx context.Context, mc *MachineConfig) error
	StartSSHAuthService(ctx context.Context, mc *MachineConfig) error
	StartTimeSyncService(ctx context.Context, mc *MachineConfig) error
	GetVMState() *VMState
}

func (mc *MachineConfig) PodmanAPISocketHost() string {
	// io.NewDir(mc.Dirs.SocksDir).AppendFile("podman-api.sock").
	return io2.NewFile(mc.Dirs.SocksDir).AppendFile("podman-api.sock").GetPath()
}

// MakeDirs make workspace directories for vm, include logs, config, socks, data dir
func (mc *MachineConfig) MakeDirs() error {
	if err := os.MkdirAll(mc.Dirs.LogsDir, os.ModePerm); err != nil {
		return err //nolint:wrapcheck
	}

	if err := os.MkdirAll(mc.Dirs.ConfigDir, os.ModePerm); err != nil {
		return err //nolint:wrapcheck
	}

	if err := os.MkdirAll(mc.Dirs.SocksDir, os.ModePerm); err != nil {
		return err //nolint:wrapcheck
	}

	if err := os.MkdirAll(mc.Dirs.DataDir, os.ModePerm); err != nil {
		return err //nolint:wrapcheck
	}

	return os.MkdirAll(mc.Dirs.PidsDir, os.ModePerm) //nolint:wrapcheck
}

func (mc *MachineConfig) CreateSSHKey() error {
	privateKeyFile := io2.NewFile(mc.SSH.PrivateKey)
	if err := privateKeyFile.Delete(true); err != nil {
		return fmt.Errorf("delete ssh private key err: %w", err)
	}

	publicKeyFile := io2.NewFile(fmt.Sprintf("%s.pub", mc.SSH.PrivateKey))

	if err := publicKeyFile.Delete(true); err != nil {
		return fmt.Errorf("delete ssh public key err: %w", err)
	}

	var sshCommand = []string{"ssh-keygen", "-N", "", "-t", "ed25519", "-f"}
	args := append(append([]string{}, sshCommand[1:]...), mc.SSH.PrivateKey)
	cmd := exec.Command(sshCommand[0], args...)
	logrus.Infof("full cmdline: %q", cmd.Args)

	return cmd.Run() //nolint:wrapcheck
}

// GetNetworkStackEndpoint return the unix socket path for network stack endpoint which provided by gvproxy.
// the NetworkStackEndpoint provides the network stack for vm
func (mc *MachineConfig) GetNetworkStackEndpoint() string {
	return io2.NewFile(mc.Dirs.SocksDir).AppendFile(define.GvProxyEndPoint).GetPath()
}

func (mc *MachineConfig) GetSSHPort() error {
	if port.IsListening(mc.SSH.Port) {
		logrus.Warnf("%d not available, try to allocate a free port for ssh", mc.SSH.Port)
		p, err := port.GetFree()
		if err != nil {
			return fmt.Errorf("failed to get free port: %w", err)
		}
		logrus.Infof("get free port: %d", p)
		mc.SSH.Port = p
	}

	return nil
}

type MachineDirs struct {
	ConfigDir string `json:"configDir" validate:"required"`
	DataDir   string `json:"dataDir"   validate:"required"`
	PidsDir   string `json:"pidsDir"   validate:"required"`
	LogsDir   string `json:"logsDir"   validate:"required"`
	SocksDir  string `json:"socksDir"  validate:"required"`
}

type MachineConfig struct {
	Dirs         MachineDirs     `json:"dirs"                validate:"required"`
	VMName       string          `json:"name"                validate:"required"`
	Bootable     Bootable        `json:"bootable"            validate:"required"`
	DataDisk     DataDisk        `json:"dataDisk"            validate:"required"`
	ConfigFile   string          `json:"configFile"          validate:"required"`
	Resources    ResourceConfig  `json:"resources"`
	Mounts       []volumes.Mount `json:"mounts"`
	SSH          SSHConfig       `json:"ssh"                 validate:"required"`
	ReportURL    string          `json:"reportURL,omitempty"`
	PodmanSocks  podmanSocks     `json:"podmanSocks"         validate:"required"`
	PIDFiles     pidFiles        `json:"pidFiles"`
	SSHAuthSocks SSHAuthSocks    `json:"sshAuthSocks"        validate:"required"`

	// RestAPISocks is the socks for rest api, it is used by appliance to connect to query the status of vm
	// exec cmdline in vm etc...
	RestAPISocks string `json:"restAPISocks" validate:"required"`
	KrunKitBin   string `json:"krunKitBin"   validate:"required"`
	VFKitBin     string `json:"vfKitBin"     validate:"required"`
	GVProxyBin   string `json:"gvProxyBin"   validate:"required"`
}

type SSHAuthSocks struct {
	LocalSocks  string `json:"localSocks"  validate:"required"`
	RemoteSocks string `json:"remoteSocks" validate:"required"`
}

// gvproxy will forward the connect from host (InHost) to guest(InGuest)
type podmanSocks struct {
	// podman api socks in host
	InHost string `json:"inHost" validate:"required"`
	// podman api socks in guest
	InGuest string `json:"inGuest" validate:"required"`
}

// TODO: kill old pid before launch new process
// pidFiles contains the pid files for gvproxy, krunKit and vfKit
type pidFiles struct {
	GvproxyPidFile string `json:"gvproxyPidFile"`
	KrunKitPidFile string `json:"krunKitPidFile"`
	VFKitPidFile   string `json:"vfKitPidFile"`
}

type Bootable struct {
	Path    string `json:"path"    validate:"required"`
	Version string `json:"version" validate:"required"`
}

type DataDisk struct {
	Path    string `json:"path"    validate:"required"`
	Version string `json:"version" validate:"required"`
}

// SSHConfig contains remote access information for SSH
type SSHConfig struct {
	PrivateKey     string `json:"identityPath"   validate:"required"`
	PublicKey      string `json:"publicKey"      validate:"required"`
	Port           int    `json:"port"           validate:"required"`
	RemoteUsername string `json:"remoteUsername" validate:"required"`
}

// NewMachineConfig initializes and returns a new MachineConfig object using the provided VMOpts configuration.
func NewMachineConfig(opts VMOpts) *MachineConfig {
	mc := new(MachineConfig)
	mc.VMName = opts.VMName

	space := workspace.GetWorkspace()

	mc.Dirs.ConfigDir = filepath.Join(space, define.ConfigPrefixDir)
	mc.Dirs.DataDir = filepath.Join(space, define.DataPrefixDir)
	mc.Dirs.LogsDir = filepath.Join(space, define.LogPrefixDir)
	mc.Dirs.SocksDir = filepath.Join(space, define.SocksPrefixDir)
	mc.Dirs.PidsDir = filepath.Join(space, define.PidsPrefixDir)

	mc.ConfigFile = filepath.Join(mc.Dirs.ConfigDir, fmt.Sprintf("%s.json", opts.VMName))
	mc.Resources = ResourceConfig{
		CPUs:           opts.CPUs,
		DataDiskSizeGB: define.DataDiskSizeInGB,
		MemoryInMB:     opts.MemoryInMiB,
	}

	mc.SSH = SSHConfig{
		PrivateKey:     filepath.Join(mc.Dirs.DataDir, define.SSHKey),
		PublicKey:      filepath.Join(mc.Dirs.DataDir, fmt.Sprintf("%s.pub", define.SSHKey)),
		Port:           define.DefaultSSHPort,
		RemoteUsername: define.DefaultUserInVM,
	}

	mc.PodmanSocks.InHost = filepath.Join(mc.Dirs.SocksDir, define.PodmanHostSocksName)
	mc.PodmanSocks.InGuest = define.PodmanGuestSocks

	mc.RestAPISocks = filepath.Join(mc.Dirs.SocksDir, define.RESTAPIEndpointName)

	mc.Bootable.Version = opts.BootVersion
	mc.Bootable.Path = filepath.Join(mc.Dirs.DataDir, fmt.Sprintf("%s.img", mc.VMName))

	mc.DataDisk.Version = opts.DataVersion
	mc.DataDisk.Path = filepath.Join(mc.Dirs.DataDir, "data.img")

	mc.Mounts = volumes.CmdLineVolumesToMounts(opts.Volumes)

	mc.ReportURL = opts.ReportURL

	// Set PIDFiles
	mc.PIDFiles.GvproxyPidFile = filepath.Join(mc.Dirs.PidsDir, define.GvProxyPidName)
	mc.PIDFiles.KrunKitPidFile = filepath.Join(mc.Dirs.PidsDir, define.KrunkitPidFile)
	mc.PIDFiles.VFKitPidFile = filepath.Join(mc.Dirs.PidsDir, define.VFkitPidFile)

	// Set SSHAuthSocks
	mc.SSHAuthSocks.LocalSocks = filepath.Join(mc.Dirs.SocksDir, define.SSHAuthLocalSockName)
	mc.SSHAuthSocks.RemoteSocks = "/opt/ssh_auth/oo-ssh-agent.sock"

	return mc
}

var (
	ErrInvalidJsonFormat = errors.New("invalid json format")
)

// LoadMachineFromFQPath loads a machine config from the given path and validates it
// this function must testable
func LoadMachineFromFQPath(p string) (*MachineConfig, error) {
	mc := new(MachineConfig)
	f := io2.NewFile(p)

	b, err := f.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read machine config: %w", err)
	}

	if err = json.Unmarshal(b, mc); err != nil {
		logrus.Errorf("failed to unmarshal JSON: %v", err)
		return nil, ErrInvalidJsonFormat
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(mc); err != nil {
		logrus.Errorf("invalid JSON struct fail: %v", err)
		return nil, ErrInvalidJsonFormat
	}

	return mc, nil
}

// write is a non-locking way to write the machine configuration file to disk
func (mc *MachineConfig) Write() error {
	if mc.ConfigFile == "" {
		return fmt.Errorf("no configuration file associated with vm %q", mc.VMName)
	}
	b, err := json.Marshal(mc)
	if err != nil {
		return fmt.Errorf("failed to marshal machine config: %w", err)
	}
	return ioutils.AtomicWriteFile(mc.ConfigFile, b, define.DefaultFilePerm) //nolint:wrapcheck
}
