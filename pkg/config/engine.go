//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package config

import (
	"github.com/spf13/pflag"
)

type OvmConfig struct {
	*pflag.FlagSet
	ContainersConfDefaultsRO *Config // The read-only! defaults configure
}
