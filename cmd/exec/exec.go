package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"strings"
)

const (
	BAUK_LOG_LEVEL = "BAUK_LOG_LEVEL"
	BAUK_HOST      = "BAUK_HOST"
)

var (
	defaultHost = "192.168.127.254:5321"
)

func init() {
	if host := os.Getenv(BAUK_HOST); host != "" {
		defaultHost = host
	}

	if level := os.Getenv(BAUK_LOG_LEVEL); level != "" {
		switch level {
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
		log.Fatalf("Failed to create session: %s", err)
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
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// 等待命令执行完成
	if err := session.Wait(); err != nil {
		logrus.Fatalf("Command finished with error: %s", err)
		os.Exit(err.(*ssh.ExitError).ExitStatus())
	}
	fmt.Println("Command executed successfully")
}
