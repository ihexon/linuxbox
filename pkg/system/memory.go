//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import (
	"fmt"

	"github.com/containers/common/pkg/strongunits"
	"github.com/shirou/gopsutil/v3/mem"
)

// checkMaxMemory gets the total system memory and compares it to the variable.  if the variable
// is larger than the total memory, it returns an error
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
