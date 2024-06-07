//go:build generate || windows

package internal

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/google/deck"
	"golang.org/x/sys/windows"
)

var (
	modDismAPI = windows.NewLazySystemDLL("DismAPI.dll")

	procDismAddCapability    = modDismAPI.NewProc("DismAddCapability")
	procDismAddDriver        = modDismAPI.NewProc("DismAddDriver")
	procDismAddPackage       = modDismAPI.NewProc("DismAddPackage")
	procDismApplyUnattend    = modDismAPI.NewProc("DismApplyUnattend")
	procDismCloseSession     = modDismAPI.NewProc("DismCloseSession")
	procDismDisableFeature   = modDismAPI.NewProc("DismDisableFeature")
	procDismEnableFeature    = modDismAPI.NewProc("DismEnableFeature")
	procDismInitialize       = modDismAPI.NewProc("DismInitialize")
	procDismOpenSession      = modDismAPI.NewProc("DismOpenSession")
	procDismRemoveCapability = modDismAPI.NewProc("DismRemoveCapability")
	procDismRemoveDriver     = modDismAPI.NewProc("DismRemoveDriver")
	procDismRemovePackage    = modDismAPI.NewProc("DismRemovePackage")
	procDismShutdown         = modDismAPI.NewProc("DismShutdown")
	procDismGetFeatures      = modDismAPI.NewProc("DismGetFeatures") // TODO
)

var _ unsafe.Pointer

const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// https://github.com/ziglang/zig/blob/6a65561e3e5f82f126ec4795e5cd9c07392b457b/lib/libc/include/any-windows-any/dismapi.h
const (
	DISM_ONLINE_IMAGE = "DISM_{53BFAE52-B167-4E2F-A258-0A37B57FF845}"

	DISM_MOUNT_READWRITE       = 0x00000000
	DISM_MOUNT_READONLY        = 0x00000001
	DISM_MOUNT_OPTIMIZE        = 0x00000002
	DISM_MOUNT_CHECK_INTEGRITY = 0x00000004

	DISMAPI_S_RELOAD_IMAGE_SESSION_REQUIRED syscall.Errno = 0x00000001
)

// DismPackageIdentifier specifies whether a package is identified by name or by file path.
type DismPackageIdentifier uint32

const (
	// DismPackageNone indicates that no package is specified.
	DismPackageNone DismPackageIdentifier = iota
	// DismPackageName indicates that the package is identified by its name.
	DismPackageName
	// DismPackagePath indicates that the package is specified by its path.
	DismPackagePath
)

// Session holds a dism session. You must call Close() to free up the session upon completion.
type Session struct {
	Handle         *uintptr
	imagePath      string
	optWindowsDir  string
	optSystemDrive string
}

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

func DismAddCapability(Session uintptr, Name *uint16, LimitAccess bool, SourcePaths **uint16, SourcePathCount uint32, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) {
	var _p0 uint32
	if LimitAccess {
		_p0 = 1
	}
	r0, _, _ := syscall.Syscall9(procDismAddCapability.Addr(), 8, uintptr(Session), uintptr(unsafe.Pointer(Name)), uintptr(_p0), uintptr(unsafe.Pointer(SourcePaths)), uintptr(SourcePathCount), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData), 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismAddDriver(Session uintptr, DriverPath *uint16, ForceUnsigned bool) (e error) {
	var _p0 uint32
	if ForceUnsigned {
		_p0 = 1
	}
	r0, _, _ := syscall.Syscall(procDismAddDriver.Addr(), 3, uintptr(Session), uintptr(unsafe.Pointer(DriverPath)), uintptr(_p0))
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismAddPackage(Session uintptr, PackagePath *uint16, IgnoreCheck bool, PreventPending bool, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) {
	var _p0 uint32
	if IgnoreCheck {
		_p0 = 1
	}
	var _p1 uint32
	if PreventPending {
		_p1 = 1
	}
	r0, _, _ := syscall.Syscall9(procDismAddPackage.Addr(), 7, uintptr(Session), uintptr(unsafe.Pointer(PackagePath)), uintptr(_p0), uintptr(_p1), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData), 0, 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismApplyUnattend(Session uintptr, UnattendFile *uint16, SingleSession bool) (e error) {
	var _p0 uint32
	if SingleSession {
		_p0 = 1
	}
	r0, _, _ := syscall.Syscall(procDismApplyUnattend.Addr(), 3, uintptr(Session), uintptr(unsafe.Pointer(UnattendFile)), uintptr(_p0))
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismCloseSession(Session uintptr) (e error) {
	r0, _, _ := syscall.Syscall(procDismCloseSession.Addr(), 1, uintptr(Session), 0, 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismDisableFeature(Session uintptr, FeatureName *uint16, PackageName *uint16, RemovePayload bool, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) {
	var _p0 uint32
	if RemovePayload {
		_p0 = 1
	}
	r0, _, _ := syscall.Syscall9(procDismDisableFeature.Addr(), 7, uintptr(Session), uintptr(unsafe.Pointer(FeatureName)), uintptr(unsafe.Pointer(PackageName)), uintptr(_p0), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData), 0, 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismEnableFeature(Session uintptr,
	FeatureName *uint16,
	Identifier *uint16,
	PackageIdentifier *DismPackageIdentifier,
	LimitAccess bool,
	SourcePaths *string,
	SourcePathCount uint32,
	EnableAll bool,
	CancelEvent *windows.Handle,
	Progress unsafe.Pointer,
	UserData unsafe.Pointer) (e error) {
	var _p0 uint32
	if LimitAccess {
		_p0 = 1
	}
	var _p1 uint32
	if EnableAll {
		_p1 = 1
	}
	r0, _, _ := syscall.Syscall12(procDismEnableFeature.Addr(), 11, uintptr(Session), uintptr(unsafe.Pointer(FeatureName)), uintptr(unsafe.Pointer(Identifier)), uintptr(unsafe.Pointer(PackageIdentifier)), uintptr(_p0), uintptr(unsafe.Pointer(SourcePaths)), uintptr(SourcePathCount), uintptr(_p1), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData), 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismInitialize(LogLevel DismLogLevel, LogFilePath *uint16, ScratchDirectory *uint16) (e error) {
	r0, _, _ := syscall.Syscall(procDismInitialize.Addr(), 3, uintptr(LogLevel), uintptr(unsafe.Pointer(LogFilePath)), uintptr(unsafe.Pointer(ScratchDirectory)))
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismOpenSession(ImagePath *uint16, WindowsDirectory *uint16, SystemDrive *uint16, Session *uintptr) (e error) {
	r0, _, _ := syscall.Syscall6(procDismOpenSession.Addr(), 4, uintptr(unsafe.Pointer(ImagePath)), uintptr(unsafe.Pointer(WindowsDirectory)), uintptr(unsafe.Pointer(SystemDrive)), uintptr(unsafe.Pointer(Session)), 0, 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismRemoveCapability(Session uintptr, Name *uint16, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) {
	r0, _, _ := syscall.Syscall6(procDismRemoveCapability.Addr(), 5, uintptr(Session), uintptr(unsafe.Pointer(Name)), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData), 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismRemoveDriver(Session uintptr, DriverPath *uint16) (e error) {
	r0, _, _ := syscall.Syscall(procDismRemoveDriver.Addr(), 2, uintptr(Session), uintptr(unsafe.Pointer(DriverPath)), 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismRemovePackage(Session uintptr, Identifier *uint16, PackageIdentifier *DismPackageIdentifier, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) {
	r0, _, _ := syscall.Syscall6(procDismRemovePackage.Addr(), 6, uintptr(Session), uintptr(unsafe.Pointer(Identifier)), uintptr(unsafe.Pointer(PackageIdentifier)), uintptr(unsafe.Pointer(CancelEvent)), uintptr(Progress), uintptr(UserData))
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func DismShutdown() (e error) {
	r0, _, _ := syscall.Syscall(procDismShutdown.Addr(), 0, 0, 0, 0)
	if r0 != 0 {
		e = syscall.Errno(r0)
	}
	return
}

func (s Session) AddCapability(
	name string,
	limitAccess bool,
	sourcePaths string,
	sourcePathsCount uint32,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	var sp **uint16
	if p := StringToPtrOrNil(sourcePaths); p != nil {
		sp = &p
	}
	return s.checkError(DismAddCapability(*s.Handle, StringToPtrOrNil(name), limitAccess, sp, sourcePathsCount, cancelEvent, progressCallback, nil))
}

func (s Session) AddPackage(
	packagePath string,
	ignoreCheck bool,
	preventPending bool,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismAddPackage(*s.Handle, StringToPtrOrNil(packagePath), ignoreCheck, preventPending, cancelEvent, progressCallback, nil))
}

func (s Session) DisableFeature(
	feature string,
	optPackageName string,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismDisableFeature(*s.Handle, StringToPtrOrNil(feature), StringToPtrOrNil(optPackageName), false, cancelEvent, progressCallback, nil))
}

func (s Session) EnableFeature(
	feature string,
	optIdentifier string,
	optPackageIdentifier *DismPackageIdentifier,
	enableAll bool,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismEnableFeature(*s.Handle, StringToPtrOrNil(feature), StringToPtrOrNil(optIdentifier), optPackageIdentifier, false, nil, 0, enableAll, cancelEvent, progressCallback, nil))
}

func (s Session) RemoveCapability(
	name string,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismRemoveCapability(*s.Handle, StringToPtrOrNil(name), cancelEvent, progressCallback, nil))
}

func (s Session) RemovePackage(
	identifier string,
	packageIdentifier *DismPackageIdentifier,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismRemovePackage(*s.Handle, StringToPtrOrNil(identifier), packageIdentifier, cancelEvent, progressCallback, nil))
}

func (s Session) Close() error {
	if err := DismCloseSession(*s.Handle); err != nil {
		return err
	}
	return DismShutdown()
}

func (s Session) checkError(err error) error {
	if err == DISMAPI_S_RELOAD_IMAGE_SESSION_REQUIRED {
		if err := DismCloseSession(*s.Handle); err != nil {
			deck.Warningf("Closing session before reloading failed: %s", err.Error())
		}

		if err := DismOpenSession(StringToPtrOrNil(s.imagePath), StringToPtrOrNil(s.optWindowsDir), StringToPtrOrNil(s.optSystemDrive), s.Handle); err != nil {
			return fmt.Errorf("reloading session: %w", err)
		}
		deck.Infof("Reloaded image session as requested by DISM API")

		return nil
	}

	return err
}

type DismLogLevel uint32

const (
	// DismLogErrors logs only errors.
	DismLogErrors DismLogLevel = 0
	// DismLogErrorsWarnings logs errors and warnings.
	DismLogErrorsWarnings DismLogLevel = 1
	// DismLogErrorsWarningsInfo logs errors, warnings, and additional information.
	DismLogErrorsWarningsInfo DismLogLevel = 2
)

func OpenSession(imagePath, optWindowsDir, optSystemDrive string, logLevel DismLogLevel, optLogFilePath, optScratchDir string) (Session, error) {
	var handleVal uintptr
	s := Session{
		Handle:         &handleVal,
		imagePath:      imagePath,
		optWindowsDir:  optWindowsDir,
		optSystemDrive: optSystemDrive,
	}

	if err := DismInitialize(logLevel, StringToPtrOrNil(optLogFilePath), StringToPtrOrNil(optScratchDir)); err != nil {
		return s, fmt.Errorf("DismInitialize: %w", err)
	}

	if err := DismOpenSession(StringToPtrOrNil(imagePath), StringToPtrOrNil(optWindowsDir), StringToPtrOrNil(optSystemDrive), s.Handle); err != nil {
		return s, fmt.Errorf("DismOpenSession: %w", err)
	}

	return s, nil
}
