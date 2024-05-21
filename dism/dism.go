//go:build generate || windows

package dism

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type (
	DismLogLevel          uint32
	DismPackageIdentifier uint32
)

type Session struct {
	Handle         *uint32
	imagePath      string
	optWindowsDir  string
	optSystemDrive string
}

const (
	DISM_ONLINE_IMAGE                                     = "DISM_{53BFAE52-B167-4E2F-A258-0A37B57FF845}"
	DISMAPI_S_RELOAD_IMAGE_SESSION_REQUIRED syscall.Errno = 0x00000001

	// DismLogErrors logs only errors.
	DismLogErrors DismLogLevel = 0
	// DismLogErrorsWarnings logs errors and warnings.
	DismLogErrorsWarnings DismLogLevel = 1
	// DismLogErrorsWarningsInfo logs errors, warnings, and additional information.
	DismLogErrorsWarningsInfo DismLogLevel = 2
)

func StringToPtrOrNil(in string) (out *uint16) {
	if in != "" {
		out = windows.StringToUTF16Ptr(in)
	}
	return
}

func (s Session) checkError(err error) error {

	if err == DISMAPI_S_RELOAD_IMAGE_SESSION_REQUIRED {
		if err := DismCloseSession(*s.Handle); err != nil {
			fmt.Errorf("Closing session before reloading failed: %s", err.Error())
		}

		if err := DismOpenSession(StringToPtrOrNil(s.imagePath), StringToPtrOrNil(s.optWindowsDir), StringToPtrOrNil(s.optSystemDrive), s.Handle); err != nil {
			return fmt.Errorf("Opening session before reloading failed: %s", err.Error())
		}
		fmt.Println("Reloaded image session as requested by DISM API")
		return nil
	}
	return err
}

// Close closes the session and shuts down dism. This must be called prior to exiting.
func (s Session) Close() error {
	if err := DismCloseSession(*s.Handle); err != nil {
		return err
	}
	return DismShutdown()
}

// OpenSession opens a DISM session. The session can be used for subsequent DISM calls.
//
// Don't forget to call Close() on the returned Session object.
//
// Example, modifying the online image:
//
//	dism.OpenSession(dism.DISM_ONLINE_IMAGE, "", "", dism.DismLogErrorsWarningsInfo, "", "")
//
// Ref: https://docs.microsoft.com/en-us/windows-hardware/manufacture/desktop/dism/disminitialize-function
//
// Ref: https://docs.microsoft.com/en-us/windows-hardware/manufacture/desktop/dism/dismopensession-function
func OpenSession(imagePath, optWindowsDir, optSystemDrive string, logLevel DismLogLevel, optLogFilePath, optScratchDir string) (Session, error) {

	var handleVal uint32
	session := Session{
		Handle:         &handleVal,
		imagePath:      imagePath,
		optWindowsDir:  optWindowsDir,
		optSystemDrive: optSystemDrive,
	}

	if err := DismInitialize(logLevel, StringToPtrOrNil(optLogFilePath), StringToPtrOrNil(optScratchDir)); err != nil {
		return session, fmt.Errorf("DismInitialize: %w", err)
	}

	if err := DismOpenSession(StringToPtrOrNil(imagePath), StringToPtrOrNil(""), StringToPtrOrNil(""), session.Handle); err != nil {
		return session, fmt.Errorf("DismOpenSession: %w", err)
	}
	return session, nil
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

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zdism.go dism.go
//sys DismInitialize(LogLevel DismLogLevel, LogFilePath *uint16, ScratchDirectory *uint16) (e error) = DismAPI.DismInitialize
//sys DismOpenSession(ImagePath *uint16, WindowsDirectory *uint16, SystemDrive *uint16, Session *uint32) (e error) = DismAPI.DismOpenSession
//sys DismCloseSession(Session uint32) (e error) = DismAPI.DismCloseSession
//sys DismEnableFeature(Session uint32, FeatureName *uint16, Identifier *uint16, PackageIdentifier *DismPackageIdentifier, LimitAccess bool, SourcePaths *string, SourcePathCount uint32, EnableAll bool, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) = DismAPI.DismEnableFeature
//sys DismShutdown() (e error) = DismAPI.DismShutdown
