//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"
	"log"
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
	logrus.Infof("Host:%s", defaultHost)

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

	// 连接 SSH 服务器
	logrus.Infof("Connecting to server %s", defaultHost)
	client, err := ssh.Dial("tcp", defaultHost, config)
	if err != nil {
		logrus.Fatalf("Failed to dial: %s", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		// TODO: @ihexon 这里会退出整个程序，确定么？
		log.Fatalf("Failed to create session: %s", err.Error())
	}
	defer session.Close()

	// 获取命令的输出
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout: %s", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr: %s", err)
	}

	// 开始执行命令
	if err := session.Start(str); err != nil {
		logrus.Fatalf("Failed to start command: %s", err)
		os.Exit(err.(*ssh.ExitError).ExitStatus())
	}

	// 实时输出命令执行结果
	go func() {
		_, _ = io.Copy(os.Stdout, stdout)
	}()
	go func() {
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	// 等待命令执行完成
	if err := session.Wait(); err != nil {
		logrus.Fatalf("Command finished with error: %s", err)
		os.Exit(err.(*ssh.ExitError).ExitStatus())
	}
	fmt.Println("Command executed successfully")
}
