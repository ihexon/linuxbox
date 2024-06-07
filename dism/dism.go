//go:build generate || windows

package dism

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type (
	DismLogLevel          uint32
	DismPackageIdentifier uint32
)

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

type Session struct {
	Handle         *uint32
	imagePath      string
	optWindowsDir  string
	optSystemDrive string
	optLogFilePath string
	optScratchDir  string
}

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
		optLogFilePath: optLogFilePath,
		optScratchDir:  optScratchDir,
	}

	if err := DismInitialize(logLevel, StringToPtrOrNil(session.optLogFilePath), StringToPtrOrNil(session.optScratchDir)); err != nil {
		return session, fmt.Errorf("DismInitialize: %w", err)
	}

	// If the value of WindowsDirectory is NULL, the default value of "Windows" is used.
	//
	// The WindowsDirectory parameter cannot be used when the ImagePath parameter is set to DISM_ONLINE_IMAGE.
	//
	// If SystemDrive is NULL, the default value of the drive containing the mount point is used.
	//
	//The SystemDrive parameter cannot be used when the ImagePath parameter is set to DISM_ONLINE_IMAGE.
	if err := DismOpenSession(StringToPtrOrNil(imagePath), StringToPtrOrNil(session.optWindowsDir), StringToPtrOrNil(session.optSystemDrive), session.Handle); err != nil {
		return session, fmt.Errorf("DismOpenSession: %w", err)
	}
	return session, nil
}

// EnableFeatureA : Enable a Feature
//
// Ref: https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/dism/dismenablefeature-function?view=windows-11
func (s Session) EnableFeatureA(
	feature string,
	optIdentifier string,
	optPackageIdentifier *DismPackageIdentifier,
	enableAll bool,
	cancelEvent *windows.Handle,
	progressCallback unsafe.Pointer,
) error {
	return s.checkError(DismEnableFeature(*s.Handle, StringToPtrOrNil(feature), StringToPtrOrNil(optIdentifier), optPackageIdentifier, false, nil, 0, enableAll, cancelEvent, progressCallback, nil))
}

func (s Session) EnableFeature(features []string) error {

	for _, f := range features {
		err := s.EnableFeatureA(f, "", nil, true, nil, nil)
		if err != nil {
			needReboot := errors.Is(err, windows.ERROR_SUCCESS_REBOOT_REQUIRED) || errors.Is(err, windows.ERROR_SUCCESS_RESTART_REQUIRED)
			return fmt.Errorf("err > %v, %v", err, needReboot)
			if !needReboot {
				return fmt.Errorf(err.Error())
			}
		}
	}
	return nil
}

func (s Session) GetFeatures() ([]FeatureList, error) {

	var (
		Count        uint32 = 1024
		buf                 = make([]byte, Count*((uint32)(unsafe.Sizeof(_DismFeatureA{}))))
		sizeOfStruct        = unsafe.Sizeof(_DismFeatureA{}.FeatureName) + unsafe.Sizeof(_DismFeatureA{}.State)
	)

	err := DismGetFeatures(*s.Handle, nil, nil, &buf, &Count)
	if err != nil {
		fmt.Printf("Failed to get features: %v\n", err)
		return nil, err
	}

	buf = buf[:Count*((uint32)(sizeOfStruct))]
	pbuf := &buf[0]

	featStructList := make([]FeatureList, Count)

	for i := uint32(0); i < Count; i++ {
		pDismFeature := (*_DismFeatureA)(unsafe.Pointer(uintptr(unsafe.Pointer(pbuf)) + uintptr(i*(uint32(sizeOfStruct)))))
		pfeatName := pDismFeature.FeatureName
		featStatee := pDismFeature.State
		featName := windows.UTF16PtrToString(pfeatName)
		featStructList[i] = FeatureList{
			FeatureName: featName,
			State:       featStatee,
		}
		//fmt.Printf("%d:%v\n", i, featStructList[i])
	}
	return featStructList, nil
}

type FeatureList struct {
	FeatureName string
	State       uint32
}

type _DismFeatureA struct {
	FeatureName *uint16
	State       uint32
}

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zdism.go dism.go
//sys DismInitialize(LogLevel DismLogLevel, LogFilePath *uint16, ScratchDirectory *uint16) (e error) = DismAPI.DismInitialize
//sys DismOpenSession(ImagePath *uint16, WindowsDirectory *uint16, SystemDrive *uint16, Session *uint32) (e error) = DismAPI.DismOpenSession
//sys DismCloseSession(Session uint32) (e error) = DismAPI.DismCloseSession
//sys DismEnableFeature(Session uint32, FeatureName *uint16, Identifier *uint16, PackageIdentifier *DismPackageIdentifier, LimitAccess bool, SourcePaths *string, SourcePathCount uint32, EnableAll bool, CancelEvent *windows.Handle, Progress unsafe.Pointer, UserData unsafe.Pointer) (e error) = DismAPI.DismEnableFeature
//sys DismShutdown() (e error) = DismAPI.DismShutdown
//sys DismGetFeatures(Session uint32, Identifier *uint16, PackageIdentifier *DismPackageIdentifier, Feature *[]byte, Count *uint32) (e error) = DismAPI.DismGetFeatures
