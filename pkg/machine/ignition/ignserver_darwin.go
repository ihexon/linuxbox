package ignition

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"net/url"
	"os"
)

// ServeIgnitionOverSockV2 is a block function, design to be running in go routine
func ServeIgnitionOverSockV2(cfg *define.VMFile, mc *vmconfigs.MachineConfig) error {
	unixSocksFile, err := mc.IgnitionSocket()
	if err != nil {
		return err
	}

	_url := "unix:///" + unixSocksFile.GetPath()
	listenAddr, err := url.Parse(_url)
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

	return ServeIgnitionOverSocketCommon(listenAddr, file)
}
