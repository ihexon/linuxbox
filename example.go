package main

import (
	"MyGoPj/dism"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
)

func main() {
	dismSession, err := dism.OpenSession(dism.DISM_ONLINE_IMAGE,
		"",
		"",
		dism.DismLogErrorsWarningsInfo,
		"",
		"")

	if err != nil {
		panic(err)
	}
	defer dismSession.Close()

	if err := dismSession.EnableFeature("Containers", "", nil, true, nil, nil); err != nil {
		if errors.Is(err, windows.ERROR_SUCCESS_REBOOT_REQUIRED) {
			fmt.Printf("Please reboot!")
		} else if e, ok := err.(syscall.Errno); ok && int(e) == 1 {
			fmt.Printf("error code %d with message \"%s\"", int(e), err)
			panic(err)
		} else {
			panic(err)
		}
	}
	fmt.Print("Feature enabled")
}
