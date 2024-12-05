//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func CommonSSHSilent(username, identityPath, name string, sshPort int, inputArgs []string) error {
	return commonBuiltinSSH(username, identityPath, name, sshPort, inputArgs, false, nil)
}

func commonBuiltinSSH(username, identityPath, name string, sshPort int, inputArgs []string, passOutput bool, stdin io.Reader) error {
	config, err := createConfig(username, identityPath)
	if err != nil {
		return fmt.Errorf("failed to create ssh config: %w", err)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", sshPort), config)
	if err != nil {
		return fmt.Errorf("failed to dial ssh: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	cmd := strings.Join(inputArgs, " ")
	logrus.Debugf("Running ssh command on machine %q: %s", name, cmd)
	session.Stdin = stdin
	if passOutput {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
	} else if logrus.IsLevelEnabled(logrus.DebugLevel) {
		return runSessionWithDebug(session, cmd)
	}

	return session.Run(cmd) //nolint:wrapcheck
}

func runSessionWithDebug(session *ssh.Session, cmd string) error {
	outPipe, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	errPipe, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	logOuput := func(pipe io.Reader, done chan struct{}) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			logrus.Debugf("ssh output: %s", scanner.Text())
		}
		done <- struct{}{}
	}
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start ssh command: %w", err)
	}
	completed := make(chan struct{}, 2)
	go logOuput(outPipe, completed)
	go logOuput(errPipe, completed)
	<-completed
	<-completed

	return session.Wait() //nolint:wrapcheck
}

func createConfig(user string, identityPath string) (*ssh.ClientConfig, error) {
	key, err := os.ReadFile(identityPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
