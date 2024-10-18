package registry

import (
	defconfig "bauklotze/pkg/config"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"sync"
)

type CliCommand struct {
	Command *cobra.Command
	Parent  *cobra.Command
}

var (
	ovmSync  sync.Once
	exitCode = 0
	// Commands All commands will be registin here
	Commands   []CliCommand
	ovmOptions defconfig.OvmConfig
)

func newOvmConfig() {
	defaultConfig := defconfig.New(&defconfig.Options{
		SetDefault: true, // This makes sure that following calls to config.Default() return default config
	})

	ovmOptions = defconfig.OvmConfig{ContainersConfDefaultsRO: defaultConfig}
}

func OvmInitConfig() *defconfig.OvmConfig {
	ovmSync.Do(newOvmConfig)
	return &ovmOptions
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
