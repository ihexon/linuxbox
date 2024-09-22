//go:build !drawin && !linux && windows

package wsl

import (
	"bauklotze/pkg/config"
	"bauklotze/pkg/machine/env"
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func getElevatedOutputFileName() (string, error) {
	dir, err := env.GetVMDataDir(0)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "../", "elevated-output.log"), nil
}

func getElevatedOutputFile(mode int) (*os.File, error) {
	name, err := getElevatedOutputFileName()
	if err != nil {
		return nil, err
	}
	return os.OpenFile(name, mode, 0644)
}

func getElevatedOutputFileRead() (*os.File, error) {
	return getElevatedOutputFile(os.O_RDONLY)
}

func getElevatedOutputFileWrite() (*os.File, error) {
	return getElevatedOutputFile(os.O_WRONLY | os.O_CREATE | os.O_APPEND)
}

func truncateElevatedOutputFile() error {
	name, err := getElevatedOutputFileName()
	if err != nil {
		return err
	}

	return os.Truncate(name, 0)
}

func dumpOutputFile() {
	file, err := getElevatedOutputFileRead()
	if err != nil {
		logrus.Debug("could not find elevated child output file")
		return
	}
	defer file.Close()
	_, _ = io.Copy(os.Stdout, file)
}

func launchElevate(operation string) error {
	if err := truncateElevatedOutputFile(); err != nil {
		return err
	}
	err := relaunchElevatedWait()
	if err != nil {
		if eerr, ok := err.(*ExitCodeError); ok {
			if eerr.code == ErrorSuccessRebootRequired {
				fmt.Println("Reboot is required to continue installation, please reboot at your convenience")
				return nil
			}
		}
		fmt.Fprintf(os.Stderr, "Elevated process failed with error: %v\n\n", err)
		dumpOutputFile()
		fmt.Fprintf(os.Stderr, wslInstallError, operation)
	}
	return err
}

func attemptFeatureInstall(reExec, admin bool) error {
	if reExec && !admin {
		return launchElevate("install the Windows WSL Features")
	}
	log, err := getElevatedOutputFileWrite()
	if err != nil {
		return err
	}
	defer log.Close()
	if err := runCmdPassThroughTee(log, "dism", "/online", "/enable-feature",
		"/featurename:Microsoft-Windows-Subsystem-Linux", "/all", "/norestart"); isMsiError(err) {
		return fmt.Errorf("could not enable WSL Feature: %w", err)
	}

	if err = runCmdPassThroughTee(log, "dism", "/online", "/enable-feature",
		"/featurename:VirtualMachinePlatform", "/all", "/norestart"); isMsiError(err) {
		return fmt.Errorf("could not enable Virtual Machine Feature: %w", err)
	}
	log.Close()
	return nil
}

func installWslKernel() error {
	log, err := getElevatedOutputFileWrite()
	if err != nil {
		return err
	}
	defer log.Close()

	message := "Installing WSL Kernel Update"
	fmt.Println(message)
	fmt.Fprintln(log, message)

	backoff := 500 * time.Millisecond
	for i := 0; i < 5; i++ {
		err = runCmdPassThroughTee(log, FindWSL(), "--update")
		if err == nil {
			break
		}
		// In case of unusual circumstances (e.g. race with installer actions)
		// retry a few times
		message = "An error occurred attempting the WSL Kernel update, retrying..."
		fmt.Println(message)
		fmt.Fprintln(log, message)
		time.Sleep(backoff)
		backoff *= 2
	}
	if err != nil {
		return fmt.Errorf("could not install WSL Kernel: %w", err)
	}
	return nil
}

func isMsiError(err error) bool {
	if err == nil {
		return false
	}

	if eerr, ok := err.(*exec.ExitError); ok {
		switch eerr.ExitCode() {
		case 0:
			fallthrough
		case ErrorSuccessRebootInitiated:
			fallthrough
		case ErrorSuccessRebootRequired:
			return false
		}
	}

	return true
}

func isWSLExist(dist string) (bool, error) {
	return wslCheckExists(dist, false)
}

func wslCheckExists(dist string, running bool) (bool, error) {
	all, err := getAllWSLDistros(running)
	if err != nil {
		return false, err
	}

	_, exists := all[dist]
	return exists, nil
}

// setupWslProxyEnv: add environments into WSLENV list
// For example: WSLENV=HOME/w:GOPATH/l:TMPDIR/p …
//
// Ref:
// https://devblogs.microsoft.com/commandline/share-environment-vars-between-wsl-and-windows/
func setupWslProxyEnv() (hasProxy bool) {
	current, _ := os.LookupEnv("WSLENV")
	for _, key := range config.ProxyEnv {
		if value, _ := os.LookupEnv(key); len(value) < 1 {
			continue
		}

		hasProxy = true
		delim := ""
		if len(current) > 0 {
			delim = ":"
		}
		current = fmt.Sprintf("%s%s%s/u", current, delim, key)
	}
	if hasProxy {
		os.Setenv("WSLENV", current)
	}
	return
}

func getAllWSLDistros(running bool) (map[string]struct{}, error) {
	args := []string{"-l", "--quiet"}
	if running {
		args = append(args, "--running")
	}
	cmd := exec.Command("wsl.exe", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command %s %v: %w", cmd.Path, args, err)
	}

	all := make(map[string]struct{})
	scanner := bufio.NewScanner(transform.NewReader(out, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			all[fields[0]] = struct{}{}
		}
	}

	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("command %s %v failed: %w (%s)", cmd.Path, args, err, strings.TrimSpace(stderr.String()))
	}
	return all, nil
}

func runCmdPassThrough(name string, arg ...string) error {
	logrus.Infof("Running command: %s %v", name, arg)
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %v failed: %w", name, arg, err)
	}
	return nil
}

func wslPipe(input string, dist string, arg ...string) error {
	newArgs := []string{"-u", "root", "-d", dist}
	newArgs = append(newArgs, arg...)
	return pipeCmdPassThrough(FindWSL(), input, newArgs...)
}

func pipeCmdPassThrough(name string, input string, arg ...string) error {
	logrus.Debugf("Running command: %s %v", name, arg)
	cmd := exec.Command(name, arg...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %v failed: %w", name, arg, err)
	}
	return nil
}

func wslInvoke(dist string, arg ...string) error {
	newArgs := []string{"-u", "root", "-d", dist}
	newArgs = append(newArgs, arg...)
	return runCmdPassThrough(FindWSL(), newArgs...)
}

func terminateDist(dist string) error {
	cmd := exec.Command(FindWSL(), "--terminate", dist)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %s %v failed: %w (%s)", cmd.Path, cmd.Args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func unregisterDist(dist string) error {
	cmd := exec.Command(FindWSL(), "--unregister", dist)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %s %v failed: %w (%s)", cmd.Path, cmd.Args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func isWSLRunning(dist string) (bool, error) {
	return wslCheckExists(dist, true)
}

func checkAndInstallWSL(reExec bool) (bool, error) {
	admin := HasAdminRights()
	if !IsWSLFeatureEnabled() {
		return false, attemptFeatureInstall(true, admin)
	}

	return true, nil
}

// IsWSLFeatureEnabled TODO: always return false for now,need to impletion
func IsWSLFeatureEnabled() bool {
	return false
}
