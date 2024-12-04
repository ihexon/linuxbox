//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

func FindPIDByCmdline(targetArgs string) ([]int32, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var matchingPIDs []int32
	for _, proc := range procs {
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}
		if strings.Contains(cmdline, targetArgs) {
			matchingPIDs = append(matchingPIDs, proc.Pid)
		}
	}
	return matchingPIDs, nil
}
