package win32privilege

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

func setAllEnv(envs []string, is_clean bool) error {

	if is_clean == true {
		windows.Clearenv()
	}

	for _, env := range envs {
		k, v, is_ok := strings.Cut(env, "=")
		if is_ok != true {
			continue
		}
		windows.Setenv(k, v)
	}
	return nil
}

func DoAsSystem(f func() error) error {
	// LockOSThread wires the calling goroutine to its current operating system thread.
	runtime.LockOSThread()
	defer func() {
		// The RevertToSelf function terminates the impersonation of a client application.
		// Refer:  https://learn.microsoft.com/en-us/windows/win32/api/securitybaseapi/nf-securitybaseapi-reverttoself
		windows.RevertToSelf()
		// UnlockOSThread undoes an earlier call to LockOSThread.
		runtime.UnlockOSThread()
	}()
	privileges := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{
				Attributes: windows.SE_PRIVILEGE_ENABLED,
			},
		},
	}
	err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr("SeDebugPrivilege"), &privileges.Privileges[0].Luid)
	if err != nil {
		return err
	}
	err = windows.ImpersonateSelf(windows.SecurityImpersonation)
	if err != nil {
		return err
	}
	var threadToken windows.Token
	err = windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_QUERY|windows.TOKEN_ADJUST_PRIVILEGES, false, &threadToken)
	if err != nil {
		return err
	}
	tokenUser, err := threadToken.GetTokenUser()
	if err == nil && tokenUser.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		threadToken.Close()
		return f()
	}
	err = windows.AdjustTokenPrivileges(threadToken, false, &privileges, uint32(unsafe.Sizeof(privileges)), nil, nil)
	threadToken.Close()
	if err != nil {
		return err
	}

	processes, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	processEntry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	var impersonationError error
	for err = windows.Process32First(processes, &processEntry); err == nil; err = windows.Process32Next(processes, &processEntry) {
		if strings.ToLower(windows.UTF16ToString(processEntry.ExeFile[:])) != "winlogon.exe" {
			continue
		}
		winlogonProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, processEntry.ProcessID)
		if err != nil {
			impersonationError = err
			continue
		}
		var winlogonToken windows.Token
		err = windows.OpenProcessToken(winlogonProcess, windows.TOKEN_QUERY|windows.TOKEN_IMPERSONATE|windows.TOKEN_DUPLICATE, &winlogonToken)
		windows.CloseHandle(winlogonProcess)
		if err != nil {
			continue
		}
		tokenUser, err := winlogonToken.GetTokenUser()
		if err != nil || !tokenUser.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
			winlogonToken.Close()
			continue
		}
		windows.CloseHandle(processes)

		var duplicatedToken windows.Token
		err = windows.DuplicateTokenEx(winlogonToken, 0, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &duplicatedToken)
		windows.CloseHandle(winlogonProcess)
		if err != nil {
			return err
		}
		newEnv, err := duplicatedToken.Environ(false)
		if err != nil {
			duplicatedToken.Close()
			return err
		}
		currentEnv := os.Environ()
		err = windows.SetThreadToken(nil, duplicatedToken)
		duplicatedToken.Close()
		if err != nil {
			return err
		}
		setAllEnv(newEnv, true)
		err = f()
		setAllEnv(currentEnv, true)
		return err
	}
	windows.CloseHandle(processes)
	if impersonationError != nil {
		return impersonationError
	}
	return errors.New("unable to find winlogon.exe process")
}

func DoAsService(serviceName string, f func() error) error {
	scm, err := mgr.Connect()
	if err != nil {
		return err
	}
	service, err := scm.OpenService(serviceName)
	scm.Disconnect()
	if err != nil {
		return err
	}
	status, err := service.Query()
	service.Close()
	if err != nil {
		return err
	}
	if status.ProcessId == 0 {
		return errors.New("service is not running")
	}
	return DoAsSystem(func() error {
		serviceProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, status.ProcessId)
		if err != nil {
			return err
		}
		var serviceToken windows.Token
		err = windows.OpenProcessToken(serviceProcess, windows.TOKEN_IMPERSONATE|windows.TOKEN_DUPLICATE, &serviceToken)
		windows.CloseHandle(serviceProcess)
		if err != nil {
			return err
		}
		var duplicatedToken windows.Token
		err = windows.DuplicateTokenEx(serviceToken, 0, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &duplicatedToken)
		serviceToken.Close()
		if err != nil {
			return err
		}
		newEnv, err := duplicatedToken.Environ(false)
		if err != nil {
			duplicatedToken.Close()
			return err
		}
		currentEnv := os.Environ()
		err = windows.SetThreadToken(nil, duplicatedToken)
		duplicatedToken.Close()
		if err != nil {
			return err
		}
		setAllEnv(newEnv, true)
		err = f()
		setAllEnv(currentEnv, true)
		return err
	})
}
