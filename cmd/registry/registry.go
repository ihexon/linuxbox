package registry

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/domain/entities"
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
	// Commands All commands will be registin here
	Commands      []CliCommand
	podmanOptions entities.OvmConfig
)

func OvmConfig() *entities.OvmConfig {
	podmanSync.Do(newPodmanConfig)
	return &podmanOptions
}

func newPodmanConfig() {
	defaultConfig, err := config.New(&config.Options{
		SetDefault: true, // This makes sure that following calls to config.Default() return this config
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain podman configuration: %v\n", err)
		os.Exit(1)
	}

	podmanOptions = entities.OvmConfig{ContainersConfRW: &config.Config{}, ContainersConfDefaultsRO: defaultConfig}

}
