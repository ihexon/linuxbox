package system

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"strings"
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
