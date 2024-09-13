package registry

import (
	defconfig "bauklotze/pkg/config"
	"bauklotze/pkg/config/domain/entities"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

type CliCommand struct {
	Command *cobra.Command
	Parent  *cobra.Command
}

var (
	podmanSync sync.Once
	exitCode   = 0
	// Commands All commands will be registin here
	Commands      []CliCommand
	podmanOptions entities.OvmConfig
)

func newPodmanConfig() {
	defaultConfig, err := defconfig.New(&defconfig.Options{
		SetDefault: true, // This makes sure that following calls to config.Default() return default config
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain ovm configuration: %v\n", err)
		os.Exit(1)
	}

	podmanOptions = entities.OvmConfig{ContainersConfRW: &defconfig.Config{}, ContainersConfDefaultsRO: defaultConfig}
}

func OvmInitConfig() *entities.OvmConfig {
	podmanSync.Do(newPodmanConfig)
	return &podmanOptions
}

func SetExitCode(code int) {
	exitCode = code
}

func GetExitCode() int {
	return exitCode
}
