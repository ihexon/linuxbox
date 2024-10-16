package registry

import (
	defconfig "bauklotze/pkg/config"
	"fmt"
	jsoniter "github.com/json-iterator/go"
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
	podmanOptions defconfig.OvmConfig
)

func newPodmanConfig() {
	defaultConfig, err := defconfig.New(&defconfig.Options{
		SetDefault: true, // This makes sure that following calls to config.Default() return default config
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain ovm configuration: %v\n", err)
		os.Exit(1)
	}

	podmanOptions = defconfig.OvmConfig{ContainersConfDefaultsRO: defaultConfig}
}

func OvmInitConfig() *defconfig.OvmConfig {
	podmanSync.Do(newPodmanConfig)
	return &podmanOptions
}

func SetExitCode(code int) {
	exitCode = code
}

func GetExitCode() int {
	return exitCode
}

var (
	json     jsoniter.API
	jsonSync sync.Once
)

// JSONLibrary provides a "encoding/json" compatible API
func JSONLibrary() jsoniter.API {
	jsonSync.Do(func() {
		json = jsoniter.ConfigCompatibleWithStandardLibrary
	})
	return json
}
