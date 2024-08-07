package wsl

import (
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"os/exec"
)

func reboot() error {

	message := "To continue the process of enabling WSL, the system needs to reboot. " +
		"Alternatively, you can cancel and reboot manually\n\n" +
		"After rebooting, please relaunch ovm and continue installing."

	if MessageBox(message, "OVMachine", false) != 1 {
		fmt.Println("Reboot is required to continue installation, please reboot at your convenience")
		os.Exit(ErrorSuccessRebootRequired)
		return nil
	}

	if err := winio.RunWithPrivilege(rebootPrivilege, func() error {
		if err := windows.ExitWindowsEx(rebootFlags, rebootReason); err != nil {
			return fmt.Errorf("execute ExitWindowsEx to reboot system failed: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cannot reboot system: %w", err)
	}

	return nil
}

func runCmdPassThroughTee(out io.Writer, name string, arg ...string) error {
	logrus.Debugf("Running command: %s %v", name, arg)

	// TODO - Perhaps improve this with a conpty pseudo console so that
	//        dism installer text bars mirror console behavior (redraw)
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, out)
	cmd.Stderr = io.MultiWriter(os.Stderr, out)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %v failed: %w", name, arg, err)
	}
	return nil
}
