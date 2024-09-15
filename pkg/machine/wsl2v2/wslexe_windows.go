package wsl2v2

import (
	"bauklotze/pkg/machine/wsl2v2/internal/state"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Backend struct{}

var (
	ErrNotExist   = errors.New("distro does not exist")
	errWslTimeout = errors.New("wsl.exe did not respond: consider restarting wslservice.exe")
)

// wslExe is a helper function to run wsl.exe with the given arguments.
// It returns the stdout, or an error containing both stdout and stderr.
func wslExe(ctx context.Context, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "wsl.exe", args...)

	// Avoid output encoding issues (WSL uses UTF-16 by default)
	cmd.Env = append(os.Environ(), "WSL_UTF8=1")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), nil
	}

	out := stdout.String()
	e := stderr.String()

	if strings.Contains(out, "Wsl/Service/WSL_E_DISTRO_NOT_FOUND") {
		return nil, ErrNotExist
	}

	if strings.Contains(e, "Wsl/Service/WSL_E_DISTRO_NOT_FOUND") {
		return nil, ErrNotExist
	}

	return nil, fmt.Errorf("%v. Stdout: %s. Stderr: %s", err, out, e)
}

func (Backend) State(distributionName string) (s state.State, err error) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), 5*time.Second, errWslTimeout)
	defer cancel()

	out, err := wslExe(ctx, "--list", "--all", "--verbose")
	if err != nil {
		return s, fmt.Errorf("could not get states of distros: %w", err)
	}

	/*
		Sample output:
		   NAME           STATE           VERSION
		 * Ubuntu         Stopped         2
		   Ubuntu-Preview Running         2
	*/

	sc := bufio.NewScanner(bytes.NewReader(out))
	var headerSkipped bool
	for sc.Scan() {
		if !headerSkipped {
			headerSkipped = true
			continue
		}

		data := strings.Fields(sc.Text())
		if len(data) == 4 {
			// default distro, ignoring leading asterisk
			data = data[1:]
		}

		if data[0] == distributionName {
			return state.NewFromString(data[1])
		}
	}

	return state.NotRegistered, nil
}
