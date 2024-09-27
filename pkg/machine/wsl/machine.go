//go:build !drawin && !linux && windows

package wsl

import (
	"bauklotze/pkg/config"
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"os"
	"os/exec"
	"strings"
)

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

// wslPipe : echo "input" | wsl -u root -d "$dist"
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

func mountBareVHDX(vhdxDisk string) error {
	logrus.Infof("map %s to all wsl2 distro", vhdxDisk)
	return pipeCmdPassThrough(FindWSL(), "--mount", "--bare", "--vhd", vhdxDisk)
}

func unmountBareVHDX(vhdxDisk string) error {
	logrus.Infof("unmap %s to all wsl2 distro", vhdxDisk)
	return pipeCmdPassThrough(FindWSL(), "--unmount", vhdxDisk)
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

// IsWSLFeatureEnabled TODO: always return false for now,need to impletion
func IsWSLFeatureEnabled() bool {
	return false
}
