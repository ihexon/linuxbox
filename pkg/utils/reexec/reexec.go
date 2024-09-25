package reexec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

func naiveSelf() string { //nolint: unused
	name := os.Args[0]
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err == nil {
			return lp
		}
	}
	// handle conversion of relative paths to absolute
	if absName, err := filepath.Abs(name); err == nil {
		return absName
	}
	// if we couldn't get absolute name, return original
	// (NOTE: Go only errors on Abs() if os.Getwd fails)
	return name
}

// Self returns the path to the current process's binary.
// Uses os.Args[0].
func Self() string {
	return naiveSelf()
}

// Command returns *exec.Cmd which has Path as current binary.
// For example if current binary is "docker" at "/usr/bin/", then cmd.Path will
// be set to "/usr/bin/docker".
func Command(args ...string) *exec.Cmd {
	cmd := exec.Command(Self())
	cmd.Args = args
	return cmd
}

// CommandContext returns *exec.Cmd which has Path as current binary.
func CommandContext(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, Self())
	cmd.Args = args
	return cmd
}
