//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package system

import "testing"

func TestFindPidByCommandLine(t *testing.T) {
	cmdline, err := FindPIDByCmdline("ovm/vm-res")
	if err != nil {
		t.Logf("error: %v", err)
	}
	for _, pid := range cmdline {
		t.Logf("pid: %d", pid)
	}
}
