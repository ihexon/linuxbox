//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	Addr string
	Port uint
	User string
	Auth []ssh.AuthMethod
}

func NewConfig(addr, user string, port uint, keyFile string) (*Config, error) {
	f, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read ssh key failed: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key")
	}

	return &Config{
		Addr: addr,
		User: user,
		Port: port,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}

func NewCmd(config *Config) *Cmd {
	return &Cmd{
		config: config,
		// send SIGKILL to stop remote process by default
		signal: ssh.SIGKILL,
	}
}
