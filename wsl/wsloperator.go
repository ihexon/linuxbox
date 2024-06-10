package wsl

import (
	"MyGoPj/shell"
	"MyGoPj/wsl/internal"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
)

func WslExec(ctx context.Context, args ...string) ([]byte, error) {
	envs := []string{"WSL_UTF8=1"}
	out, err := shell.Exec(ctx, envs, "wsl.exe", args...)
	if err == nil {
		return out, nil
	}
	fmt.Println(err.Error())
	return out, err
}

// Shutdown shutdown the wsl2 entirely
func Shutdown() error {
	_, err := WslExec(context.Background(), "--shutdown")
	if err != nil {
		return fmt.Errorf("could not shut WSL down: %w", err)
	}
	return nil
}

func Terminate(distroName string) error {
	_, err := WslExec(context.Background(), "--terminate", distroName)
	if err != nil {
		return fmt.Errorf("could not terminate distro %q: %w", distroName, err)
	}
	return nil
}

// WslState If distroName not find, return internal.NotRegistered
func WslState(distroName string) (state int, err error) {
	out, err := WslExec(context.Background(), "--list", "--all", "--verbose")
	if err != nil {
		return internal.Error, fmt.Errorf("could not get states of distros: %w", err)
	}

	{
		//   NAME                      STATE           VERSION
		//* Ubuntu                    Running         2
		//  podman-machine-default    Stopped         2
		//  ovm-test                  Stopped         2
		//  ovm                       Stopped         2
		//  ovms                      Stopped         2

		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanLines)

		line_number := 0
		for scanner.Scan() {
			line_number++
			line := scanner.Text()
			// skip the first line:
			// `NAME                      STATE           VERSION`
			if line_number == 1 {
				continue
			}
			arrayStr := strings.Fields(line)
			// Ignore symbol of `*`
			// * Ubuntu                    Running         2
			if len(arrayStr) == 4 {
				arrayStr = arrayStr[1:]
			}

			//fmt.Printf("[%d]: %s\n", line_number, arrayStr)
			if arrayStr[0] == distroName {
				switch arrayStr[1] {
				case "Running":
					return internal.Running, nil
				case "Stopped":
					return internal.Stopped, nil
				default:
					return internal.Error, fmt.Errorf("unknown state: %s", arrayStr[1])
				}
			}
		}

		if err = scanner.Err(); err != nil {
			return internal.Error, fmt.Errorf("error reading scanner: %w", err)
		}

	}
	return internal.NotRegistered, fmt.Errorf("distro %q not found", distroName)
}

// TODO: We need `Force import distro` at sometime:(
func ImportDistro(ctx context.Context, distroName, installPath, rootfs string) error {
	_, err := WslExec(ctx, "--import", distroName, installPath, rootfs)
	if err != nil {
		return fmt.Errorf("import %s failed: %v", rootfs, err)
	}
	return nil
}

func IsRegist(distroName string) (registered bool, err error) {
	statee, _ := WslState(distroName)
	if statee == internal.Error {
		return false, fmt.Errorf("unknown state: %s", distroName)
	}
	return statee != internal.NotRegistered, nil
}
