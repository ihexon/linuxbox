//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	level = "BAUK_LOG_LEVEL"
	host  = "BAUK_HOST"
)

var (
	defaultHost = "192.168.127.254:5321"
)

func init() {
	if h := os.Getenv(host); h != "" {
		defaultHost = h
	}

	if l := os.Getenv(level); l != "" {
		switch l {
		case "DEBUG":
			logrus.SetLevel(logrus.DebugLevel)
		case "INFO":
			logrus.SetLevel(logrus.InfoLevel)
		case "WARN":
			logrus.SetLevel(logrus.WarnLevel)
		case "ERROR":
			logrus.SetLevel(logrus.ErrorLevel)
		case "FATAL":
			logrus.SetLevel(logrus.FatalLevel)
		case "PANIC":
			logrus.SetLevel(logrus.PanicLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
	}
}

func main() {
	// SSH server information
	user := "ovm"
	logrus.Infof("Host:%s\n", defaultHost)

	password := "none"
	command := os.Args[1:]

	str := strings.Join(command, " ")
	if strings.TrimSpace(str) == "" || len(str) == 0 {
		logrus.Infof("Command is empty")
		return
	}

	logrus.Infof("Running [ %s ] with [ %s ]\n", command[0], command[1:])
	// Configure SSH ClientConfig
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connecting to an SSH Server
	logrus.Infof("Connecting to server %s\n", defaultHost)
	client, err := ssh.Dial("tcp", defaultHost, config)
	if err != nil {
		logrus.Fatalf("Failed to dial: %s\n", err)
	}
	defer client.Close()

	// Creating a Session
	session, err := client.NewSession()
	if err != nil {
		logrus.Fatalf("Failed to create session: %s\n", err.Error())
	}
	defer session.Close()

	// Wait for the command to complete
	stdout, err := session.StdoutPipe()
	if err != nil {
		logrus.Fatalf("Failed to get stdout: %s\n", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		logrus.Fatalf("Failed to get stderr: %s\n", err)
	}

	// Start executing commands
	if err := session.Start(str); err != nil {
		// Fatalf will call os.Exit to end the process ut we need to get the return value of the str(cmdline)
		logrus.Errorf("Failed to start command: %s\n", err)
		os.Exit(getExitCode(err)) //nolint:gocritic
	}

	// Output command execution results in real time
	go func() {
		_, _ = io.Copy(os.Stdout, stdout)
	}()
	go func() {
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	// Wait for the command to complete
	if err := session.Wait(); err != nil {
		logrus.Errorf("Command finished with error: %s\n", err)
		os.Exit(getExitCode(err))
	}
	logrus.Infof("Command executed successfully")
}

func getExitCode(err error) int {
	var sshErr *ssh.ExitError
	if ok := errors.As(err, &sshErr); ok {
		return sshErr.ExitStatus()
	}

	return 1
}
