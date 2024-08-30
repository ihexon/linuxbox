//go:build !drawin && !linux && windows

package wsl

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	once    sync.Once
	wslPath string
)

func FindWSL() string {
	// 让环境变量决定 WSL.exe 的位置
	once.Do(func() {
		wslPath = "wsl.exe"
	})
	return wslPath
}

func getLocalAppData() string {
	localapp := os.Getenv("LOCALAPPDATA")
	if localapp != "" {
		return localapp
	}

	if user := os.Getenv("USERPROFILE"); user != "" {
		return filepath.Join(user, "AppData", "Local")
	}

	return localapp
}

func SilentExec(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %v failed: %w", command, args, err)
	}
	return nil
}

func SilentExecCmd(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	return cmd
}

func MessageBox(caption, title string, fail bool) int {
	var format uint32
	if fail {
		format = windows.MB_ICONERROR
	} else {
		format = windows.MB_OKCANCEL | windows.MB_ICONINFORMATION
	}

	captionPtr, _ := syscall.UTF16PtrFromString(caption)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	ret, _ := windows.MessageBox(0, captionPtr, titlePtr, format)

	return int(ret)
}

func matchOutputLine(output io.ReadCloser, match string) bool {
	scanner := bufio.NewScanner(transform.NewReader(output, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, match) {
			return true
		}
	}
	return false
}
