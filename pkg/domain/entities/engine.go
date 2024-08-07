package entities

import (
	"bauklotze/pkg/config"
	"github.com/spf13/pflag"
)

type PodmanConfig struct {
	*pflag.FlagSet
	ContainersConfRW         *config.Config
	ContainersConfDefaultsRO *config.Config // The read-only! defaults from containers.conf.
}
