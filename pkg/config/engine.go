package config

import (
	"github.com/spf13/pflag"
)

type OvmConfig struct {
	*pflag.FlagSet
	ContainersConfRW         *Config
	ContainersConfDefaultsRO *Config // The read-only! defaults configure
}
