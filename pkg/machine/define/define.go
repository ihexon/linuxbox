package define

import (
	"os"
)

const (
	DefaultMachineName = "bugbox-machine-default"
	DefaultUserInVM    = "root"

	ConfigPrefixDir = "config"
	LogPrefixDir    = "logs"
	Libexec         = "libexec"
	DataPrefixDir   = "data"
	SocksPrefixDir  = "socks"
	PidsPrefixDir   = "pids"

	DefaultIdentityName = "sshkey"

	GvProxyBinaryName = "gvproxy"
	GvProxyPidName    = "gvproxy.pid"
	GvProxyLogName    = "gvproxy.log"
	GvProxyEndPoint   = "gvproxy.sock"

	KrunkitPidFile = "krunkit.pid"
	VFkitPidFile   = "vfkit.pid"

	VfkitBinaryName   = "vfkit"
	KrunkitBinaryName = "krunkit"

	LogFileName         = "ovm.log"
	RESTAPIEndpointName = "ovm_restapi.socks"

	LocalHostURL = "127.0.0.1"

	DefaultSSHPort = 61234

	PodmanHostSocksName = "podman-api.sock"
	PodmanGuestSocks    = "/run/podman/podman.sock"

	SSHKey = "sshkey"

	IgnMnt               = "/tmp/initfs:/tmp/initfs"
	SSHAuthLocalSockName = "oo-ssh-agent-host.sock"
)

var (
	GitCommit string
)

var (
	DataDiskSizeInGB int64       = 100
	DefaultFilePerm  os.FileMode = 0644
)
