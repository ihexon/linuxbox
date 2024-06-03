package win32privilege

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

const (
	releaseOffset      = 2
	shellExecuteOffset = 9

	cSEE_MASK_DEFAULT = 0
)

func isAdmin(token windows.Token) bool {
	builtinAdminsGroup, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}
	var checkableToken windows.Token
	err = windows.DuplicateTokenEx(token, windows.TOKEN_QUERY|windows.TOKEN_IMPERSONATE, nil, windows.SecurityIdentification, windows.TokenImpersonation, &checkableToken)
	if err != nil {
		return false
	}
	defer checkableToken.Close()
	isAdmin, err := checkableToken.IsMember(builtinAdminsGroup)
	return isAdmin && err == nil
}

func findCurrentDataTableEntry() (entry *windows.LDR_DATA_TABLE_ENTRY, err error) {
	peb := windows.RtlGetCurrentPeb()
	if peb == nil || peb.Ldr == nil {
		err = windows.ERROR_INVALID_ADDRESS
		return
	}
	for cur := peb.Ldr.InMemoryOrderModuleList.Flink; cur != &peb.Ldr.InMemoryOrderModuleList; cur = cur.Flink {
		entry = (*windows.LDR_DATA_TABLE_ENTRY)(unsafe.Pointer(uintptr(unsafe.Pointer(cur)) - unsafe.Offsetof(windows.LDR_DATA_TABLE_ENTRY{}.InMemoryOrderLinks)))
		if entry.DllBase == peb.ImageBaseAddress {
			return
		}
	}
	entry = nil
	err = windows.ERROR_OBJECT_NOT_FOUND
	return
}

func TokenIsElevatedOrElevatable(token windows.Token) bool {
	if token.IsElevated() && isAdmin(token) {
		return true
	}
	linked, err := token.GetLinkedToken()
	if err != nil {
		return false
	}
	defer linked.Close()
	return linked.IsElevated() && isAdmin(linked)
}

func ShellExecute(program, arguments, directory string, show int32) (err error) {
	var (
		program16   *uint16
		arguments16 *uint16
		directory16 *uint16
	)

	if len(program) > 0 {
		program16, _ = windows.UTF16PtrFromString(program)
	}
	if len(arguments) > 0 {
		arguments16, _ = windows.UTF16PtrFromString(arguments)
	}
	if len(directory) > 0 {
		directory16, _ = windows.UTF16PtrFromString(directory)
	}

	defer func() {
		if err != nil && program16 != nil {
			err = windows.ShellExecute(0, windows.StringToUTF16Ptr("runas"), program16, arguments16, directory16, show)
		}
	}()

	return
}
