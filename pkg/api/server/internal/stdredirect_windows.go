//go:build windows && (arm64 || amd64)

package internal

import (
	"golang.org/x/sys/windows"
	"os"
	"syscall"
)

func RedirectStdin() error {
	devNullfile, err := os.Open("NUL")
	if err != nil {
		return err
	}
	defer devNullfile.Close()

	handle := windows.Handle(devNullfile.Fd())
	if err := windows.SetStdHandle(syscall.STD_INPUT_HANDLE, handle); err != nil {
		return err
	}
	return nil
}
