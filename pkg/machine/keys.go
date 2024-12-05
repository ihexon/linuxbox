//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/containers/storage/pkg/fileutils"
)

var sshCommand = []string{"ssh-keygen", "-N", "", "-t", "ed25519", "-f"}

func CreateSSHKeys(writeLocation string) (string, error) {
	// If the SSH key already exists, hard fail
	if err := fileutils.Exists(writeLocation); err == nil {
		return "", fmt.Errorf("SSH key already exists: %s", writeLocation)
	}
	if err := os.MkdirAll(filepath.Dir(writeLocation), 0700); err != nil {
		return "", fmt.Errorf("failed to create ssh key directory: %w", err)
	}
	if err := generatekeys(writeLocation); err != nil {
		return "", fmt.Errorf("failed to generate keys: %w", err)
	}
	b, err := os.ReadFile(writeLocation + ".pub")
	if err != nil {
		return "", fmt.Errorf("failed to read ssh key: %w", err)
	}
	return strings.TrimSuffix(string(b), "\n"), nil
}

// generatekeys creates an ed25519 set of keys
func generatekeys(writeLocation string) error {
	args := append(append([]string{}, sshCommand[1:]...), writeLocation)
	cmd := exec.Command(sshCommand[0], args...)
	stdErr := &bytes.Buffer{}
	cmd.Stderr = stdErr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ssh command: %w", err)
	}
	waitErr := cmd.Wait()
	if waitErr != nil {
		return fmt.Errorf("failed to generate keys: %s: %w", strings.TrimSpace(stdErr.String()), waitErr)
	}

	return nil
}

// GetSSHKeys checks to see if there is a ssh key at the provided location.
// If not, we create the priv and pub keys. The ssh key is then returned.
func GetSSHKeys(identityPath string) (string, error) {
	if err := fileutils.Exists(identityPath); err == nil {
		// If sshkeys generated before, use it
		b, err := os.ReadFile(identityPath + ".pub")
		if err != nil {
			return "", fmt.Errorf("failed to read ssh key: %w", err)
		}
		return strings.TrimSuffix(string(b), "\n"), nil
	}
	// If ssh keys not generated before, create it
	return CreateSSHKeys(identityPath)
}
