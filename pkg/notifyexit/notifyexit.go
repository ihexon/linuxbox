//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package notifyexit

import (
	"os"

	"bauklotze/pkg/network"
)

func NotifyExit(code int) {
	network.Reporter.SendEventToOvmJs("exit", "")
	os.Exit(code)
}
