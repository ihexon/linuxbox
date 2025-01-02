//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

import (
	"fmt"
	"os/exec"
	"time"

	"bauklotze/pkg/machine/events"

	"bauklotze/pkg/httpclient"

	"github.com/sirupsen/logrus"
)

type AllCmds struct {
	Gvcmd   *exec.Cmd
	Kruncmd *exec.Cmd
}

var GlobalCmds = &AllCmds{}

func (p *AllCmds) SetGvpCmd(cmd *exec.Cmd) {
	p.Gvcmd = cmd
}

func (p *AllCmds) GetGvproxyCmd() *exec.Cmd {
	return p.Gvcmd
}

func (p *AllCmds) SetVMProviderCmd(cmd *exec.Cmd) {
	p.Kruncmd = cmd
}

func (p *AllCmds) GetVMProviderCmd() *exec.Cmd {
	return p.Kruncmd
}

func WaitAPIAndPrintInfo(sockInHost string, forwardState APIForwardingState, name string) error {
	if forwardState == NoForwarding {
		return fmt.Errorf("podman Rest API No forwarding")
	}
	err := WaitAndPingAPI("unix:///" + sockInHost)
	if err != nil {
		logrus.Error("failed to ping Podman API: ", err)
		return err
	} else {
		events.NotifyRun(events.Ready)
		fmt.Printf("Podman API forwarding listening on: %s\n", sockInHost)
	}
	return nil
}

const defaultPingTimeout = 5 * time.Second
const defaultPingInterval = 200 * time.Millisecond

func WaitAndPingAPI(sock string) error {
	client := httpclient.New().SetTransport(httpclient.CreateUnixTransport(sock))

	timeout := time.After(defaultPingTimeout)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout reached while waiting for Podman API")
		default:
			logrus.Info("Ping Podman API....")
			time.Sleep(defaultPingInterval)

			if err := client.Get("_ping"); err == nil {
				logrus.Infof("Podman ping test success")
				return nil
			}
		}
	}
}
