//go:build !drawin && !linux && windows

package wsl

import "golang.org/x/sys/windows"

const (
	ErrorSuccessRebootInitiated = 1641
	ErrorSuccessRebootRequired  = 3010
)
const (
	// ref: https://learn.microsoft.com/en-us/windows/win32/secauthz/privilege-constants#constants
	rebootPrivilege = "SeShutdownPrivilege"

	// "Application: Installation (Planned)" A planned restart or shutdown to perform application installation.
	// ref: https://learn.microsoft.com/en-us/windows/win32/shutdown/system-shutdown-reason-codes
	rebootReason = windows.SHTDN_REASON_MAJOR_APPLICATION | windows.SHTDN_REASON_MINOR_INSTALLATION | windows.SHTDN_REASON_FLAG_PLANNED

	// ref: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-exitwindowsex#parameters
	rebootFlags = windows.EWX_REBOOT | windows.EWX_RESTARTAPPS | windows.EWX_FORCEIFHUNG
)

const wslInstallError = `Could not %s. See previous output for any potential failure details.
If you can not resolve the issue, and rerunning fails, try the "wsl --install" process
outlined in the following article:

http://docs.microsoft.com/en-us/windows/wsl/install

`
