package internal

import "golang.org/x/sys/windows"

func StringToPtrOrNil(in string) (out *uint16) {
	if in != "" {
		out = windows.StringToUTF16Ptr(in)
	}
	return
}
