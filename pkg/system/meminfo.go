package system

import (
	"fmt"
	"github.com/containers/common/pkg/strongunits"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemInfo contains memory statistics of the host system.
type MemInfo struct {
	// Total usable RAM (i.e. physical RAM minus a few reserved bits and the
	// kernel binary code).
	MemTotal int64

	// Amount of free memory.
	MemFree int64

	// Total amount of swap space available.
	SwapTotal int64

	// Amount of swap space that is currently unused.
	SwapFree int64
}

func CheckMaxMemory(newMem strongunits.MiB) error {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	if total := strongunits.B(memStat.Total); strongunits.B(memStat.Total) < newMem.ToBytes() {
		return fmt.Errorf("requested amount of memory (%d MB) greater than total system memory (%d MB)", newMem, total)
	}
	return nil
}
