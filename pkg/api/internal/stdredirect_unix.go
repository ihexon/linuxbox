//go:build (darwin || linux) && (arm64 || amd64)

package internal

import (
	"golang.org/x/sys/unix"
	"os"
)

func RedirectStdin() error {
	devNullfile, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer devNullfile.Close()

	if err := unix.Dup2(int(devNullfile.Fd()), int(os.Stdin.Fd())); err != nil {
		return err
	}
	return nil
}
