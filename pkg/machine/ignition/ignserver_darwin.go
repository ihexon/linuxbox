//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

// ServeIgnitionOverSockV2 is a block function, design to be running in go routine
func ServeIgnitionOverSockV2(cfg *define.VMFile, mc *vmconfigs.MachineConfig) error {
	unixSocksFile, err := mc.IgnitionSocket()
	if err != nil {
		return fmt.Errorf("failed to get ignition socket: %w", err)
	}

	_url := "unix:///" + unixSocksFile.GetPath()
	listenAddr, err := url.Parse(_url)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	vmf, err := mc.IgnitionFile()
	if err != nil {
		return fmt.Errorf("failed to get ignition file: %w", err)
	}

	file, err := os.Open(vmf.Path)
	if err != nil {
		return fmt.Errorf("failed to open ignition file: %w", err)
	}

	return ServeIgnitionOverSocketCommon(listenAddr, file)
}

func getLocalTimeZone() (string, error) {
	tzPath, err := os.Readlink("/etc/localtime")
	if err != nil {
		return "", fmt.Errorf("failed to get local timezone: %w", err)
	}
	return strings.TrimPrefix(tzPath, "/var/db/timezone/zoneinfo"), nil
}
