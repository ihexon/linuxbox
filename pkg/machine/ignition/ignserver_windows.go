package ignition

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"net/url"
	"os"
)

// ServeIgnitionOverSock allows podman to open a small httpd instance on the vsock between the host
// and guest to inject the ignitionfile into fcos
func ServeIgnitionOverSockV2(cfg *define.VMFile, mc *vmconfigs.MachineConfig) error {
	addr, err := url.Parse(mc.IgnitionTcpListenAddr())
	if err != nil {
		return err
	}
	vmf, err := mc.IgnitionFile()
	if err != nil {
		return err
	}

	file, err := os.Open(vmf.Path)
	if err != nil {
		return err
	}

	return ServeIgnitionOverSocketCommon(addr, file)
}

func getLocalTimeZone() (string, error) {
	return "", nil
}
