package machine

import (
	"bauklotze/pkg/network"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os/exec"
	"sync"
	"time"
)

type AllCmds struct {
	Gvcmd   *exec.Cmd
	Kruncmd *exec.Cmd
	mu      sync.Mutex
}

var GlobalCmds = &AllCmds{}

func (p *AllCmds) SetGvpCmd(cmd *exec.Cmd) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Gvcmd = cmd
}

func (p *AllCmds) SetKrunCmd(cmd *exec.Cmd) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Kruncmd = cmd
}

func (p *AllCmds) GetKrunCmd() *exec.Cmd {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Kruncmd != nil {
		return p.Kruncmd
	}
	return nil
}

func (p *AllCmds) GetGvproxyCmd() *exec.Cmd {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Gvcmd != nil {
		return p.Gvcmd
	}
	return nil
}

// DO NOT BLOCK THIS FUNCTION FOR LONG TIME
func WaitAPIAndPrintInfo(sockInHost string, forwardState APIForwardingState, name string) error {
	if forwardState == NoForwarding {
		return fmt.Errorf("Podman Rest API No forwarding....")
	}
	err := WaitAndPingAPI("unix:///" + sockInHost)
	if err != nil {
		logrus.Error("failed to ping Podman API: ", err)
		return err
	} else {
		network.Reporter.SendEventToOvmJs("ready", "")
		fmt.Printf("Podman API forwarding listening on: %s\n", sockInHost)
	}
	return nil
}

func WaitAndPingAPI(sock string) error {
	connCtx, err := network.NewConnection(sock)
	if err != nil {
		return err
	}
	connCtx.UrlParameter = url.Values{}
	connCtx.Headers = http.Header{}

	timeout := time.After(5 * time.Second)
	res := &network.APIResponse{}
pingLoop:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout reached while waiting for Podman API")
		default:
			logrus.Info("Ping Podman API....")
			time.Sleep(100 * time.Microsecond)
			res, err = connCtx.DoRequest("GET", "_ping")
			if err == nil {
				_ = res.Response.Body.Close()
				logrus.Infof("Podman ping test success")
				break pingLoop
			}
		}
	}
	return err
}
