//go:build !drawin && !linux && windows

package wsl

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

type SHELLEXECUTEINFO struct {
	cbSize         uint32
	fMask          uint32
	hwnd           syscall.Handle
	lpVerb         uintptr
	lpFile         uintptr
	lpParameters   uintptr
	lpDirectory    uintptr
	nShow          int
	hInstApp       syscall.Handle
	lpIDList       uintptr
	lpClass        uintptr
	hkeyClass      syscall.Handle
	dwHotKey       uint32
	hIconOrMonitor syscall.Handle
	hProcess       syscall.Handle
}

const (
	//nolint:stylecheck
	SEE_MASK_NOCLOSEPROCESS = 0x40
	//nolint:stylecheck
	SE_ERR_ACCESSDENIED = 0x05
)

func buildCommandArgs(elevate bool) string {
	var args []string
	for _, arg := range os.Args[1:] {
		if arg != "--reexec" {
			args = append(args, syscall.EscapeArg(arg))
			if elevate && arg == "init" {
				args = append(args, "--reexec")
			}
		}
	}
	return strings.Join(args, " ")
}

func HasAdminRights() bool {
	var sid *windows.SID

	// See: https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
	if err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid); err != nil {
		logrus.Warnf("SID allocation error: %s", err)
		return false
	}
	defer func() {
		_ = windows.FreeSid(sid)
	}()

	//  From MS docs:
	// "If TokenHandle is NULL, CheckTokenMembership uses the impersonation
	//  token of the calling thread. If the thread is not impersonating,
	//  the function duplicates the thread's primary token to create an
	//  impersonation token."
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		logrus.Warnf("Token Membership Error: %s", err)
		return false
	}

	return member || token.IsElevated()
}

func wrapMaybe(err error, message string) error {
	if err != nil {
		return fmt.Errorf("%v: %w", message, err)
	}

	return errors.New(message)
}

func wrapMaybef(err error, format string, args ...interface{}) error {
	if err != nil {
		return fmt.Errorf(format+": %w", append(args, err)...)
	}

	return fmt.Errorf(format, args...)
}

type ExitCodeError struct {
	code uint
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("Process failed with exit code: %d", e.code)
}

func relaunchElevatedWait() error {
	e, _ := os.Executable()
	d, _ := os.Getwd()
	exe, _ := syscall.UTF16PtrFromString(e)
	cwd, _ := syscall.UTF16PtrFromString(d)
	arg, _ := syscall.UTF16PtrFromString(buildCommandArgs(true))
	verb, _ := syscall.UTF16PtrFromString("runas")

	shell32 := syscall.NewLazyDLL("shell32.dll")

	info := &SHELLEXECUTEINFO{
		fMask:        SEE_MASK_NOCLOSEPROCESS,
		hwnd:         0,
		lpVerb:       uintptr(unsafe.Pointer(verb)),
		lpFile:       uintptr(unsafe.Pointer(exe)),
		lpParameters: uintptr(unsafe.Pointer(arg)),
		lpDirectory:  uintptr(unsafe.Pointer(cwd)),
		nShow:        syscall.SW_SHOWNORMAL,
	}
	info.cbSize = uint32(unsafe.Sizeof(*info))
	procShellExecuteEx := shell32.NewProc("ShellExecuteExW")
	if ret, _, _ := procShellExecuteEx.Call(uintptr(unsafe.Pointer(info))); ret == 0 { // 0 = False
		err := syscall.GetLastError()
		if info.hInstApp == SE_ERR_ACCESSDENIED {
			return wrapMaybe(err, "request to elevate privileges was denied")
		}
		return wrapMaybef(err, "could not launch process, ShellEX Error = %d", info.hInstApp)
	}

	handle := info.hProcess
	defer func() {
		_ = syscall.CloseHandle(handle)
	}()

	w, err := syscall.WaitForSingleObject(handle, syscall.INFINITE)
	switch w {
	case syscall.WAIT_OBJECT_0:
		break
	case syscall.WAIT_FAILED:
		return fmt.Errorf("could not wait for process, failed: %w", err)
	default:
		return fmt.Errorf("could not wait for process, unknown error. event: %X, err: %v", w, err)
	}
	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return err
	}
	if code != 0 {
		return &ExitCodeError{uint(code)}
	}

	return nil
}
