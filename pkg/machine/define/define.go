package define

import (
	"os"

	"github.com/containers/common/pkg/strongunits"
)

const (
	DefaultMachineName = "bugbox-machine-default"
	DefaultUserInVM    = "root"

	ConfigPrefixDir  = "config"
	LogPrefixDir     = "logs"
	LibexecPrefixDir = "libexec"
	DataPrefixDir    = "data"
	TmpPrefixDir     = "tmp"

	DefaultIdentityName = "sshkey"

	GvProxyBinaryName = "gvproxy"
	GvProxyPidName    = "gvproxy.pid"
	GvProxyLogName    = "gvproxy.log"

	VfkitBinaryName   = "vfkit"
	KrunkitBinaryName = "krunkit"

	LogFileName         = "ovm.log"
	RESTAPIEndpointName = "ovm_restapi.socks"

	LocalHostURL = "127.0.0.1"

	DefaultSSHPort         = 61234
	DefaultDataImageSizeGB = 100

	PodmanGuestSocks = "/run/podman/podman.sock"
)

var (
	GitCommit string
)

var (
	DataDiskSize    strongunits.GiB = 100
	DefaultFilePerm os.FileMode     = 0644
)
