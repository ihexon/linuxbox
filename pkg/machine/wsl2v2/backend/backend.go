//go:build !drawin && !linux && windows

package backend

import (
	"bauklotze/pkg/machine/wsl2v2/internal/flags"
	"bauklotze/pkg/machine/wsl2v2/internal/state"
	"context"
	"os"
)

type Backend interface {

	// wsl.exe
	State(distributionName string) (state.State, error)
	Shutdown() error
	Terminate(distroName string) error
	SetAsDefault(distroName string) error
	Install(ctx context.Context, appxName string) error
	Import(ctx context.Context, distributionName, sourcePath, destinationPath string) error

	// Win32
	WslConfigureDistribution(distributionName string, defaultUID uint32, wslDistributionFlags flags.WslFlags) error
	WslGetDistributionConfiguration(distroName string, distributionVersion *uint8, defaultUID *uint32, wslDistributionFlags *flags.WslFlags, defaultEnvironmentVariables *map[string]string) error
	WslLaunch(distroName string, command string, useCWD bool, stdin *os.File, stdout *os.File, stderr *os.File) (*os.Process, error)
	WslLaunchInteractive(distributionName string, command string, useCurrentWorkingDirectory bool) (uint32, error)
	WslRegisterDistribution(distributionName string, tarGzFilename string) error
	WslUnregisterDistribution(distributionName string) error
}
