//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ssh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"bauklotze/pkg/machine/io"
)

var sshCommand = []string{"ssh-keygen", "-N", "", "-t", "ed25519", "-f"}

func CreateSSHKeys(f *io.VMFile) (string, error) {
	err := f.MakeBaseDir()
	if err != nil {
		return "", fmt.Errorf("failed to create ssh key directory: %w", err)
	}

	if err = generatekeys(f.GetPath()); err != nil {
		return "", fmt.Errorf("failed to generate keys: %w", err)
	}
	b, err := os.ReadFile(f.GetPath() + ".pub")
	if err != nil {
		return "", fmt.Errorf("failed to read ssh key: %w", err)
	}
	logrus.Infof("SSHPub: %s", strings.TrimSuffix(string(b), "\n"))
	return strings.TrimSuffix(string(b), "\n"), nil
}

// generatekeys creates an ed25519 set of keys
func generatekeys(writeLocation string) error {
	args := append(append([]string{}, sshCommand[1:]...), writeLocation)
	cmd := exec.Command(sshCommand[0], args...)
	stdErr := &bytes.Buffer{}
	cmd.Stderr = stdErr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ssh command:%w", err)
	}
	waitErr := cmd.Wait()
	if waitErr != nil {
		return fmt.Errorf("failed to generate keys: %s: %w", strings.TrimSpace(stdErr.String()), waitErr)
	}

	return nil
}

// GetSSHKeys checks to see if there is a ssh key at the provided location.
// If not, we create the priv and pub keys. The ssh key is then returned.
func GetSSHKeys(f *io.VMFile) (string, error) {
	if f.Exist() {
		pubF := io.VMFile{
			Path: f.GetPath() + ".pub",
		}
		logrus.Infof("SSH key already exists,read the %s", f.GetPath())
		read, err := pubF.Read()
		if err != nil {
			return "", fmt.Errorf("failed to read ssh key: %w", err)
		}
		pubKey := strings.TrimSuffix(string(read), "\n")
		logrus.Infof("Pubkey: %s", pubKey)
		return pubKey, nil
	} else {
		return CreateSSHKeys(f)
	}
}
