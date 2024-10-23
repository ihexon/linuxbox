package machine

import (
	"bauklotze/pkg/machine/system"
	"bauklotze/pkg/network"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os/exec"
	"sync"
	"time"
)

var GlobalPIDs = &AllPIDs{}

type AllPIDs struct {
	KrunkitPID *int
	GvproxyPID *int
	mu         sync.Mutex
}

func (p *AllPIDs) SetKrunkitPID(pid int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.KrunkitPID = &pid
}

func (p *AllPIDs) GetKrunkitPID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.KrunkitPID != nil {
		return *p.KrunkitPID
	}
	return 0
}

func (p *AllPIDs) SetGvproxyPID(pid int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.GvproxyPID = &pid
}

func (p *AllPIDs) GetGvproxyPID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.GvproxyPID != nil {
		return *p.GvproxyPID
	}
	return 0
}

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

func (p *AllCmds) GetKrunCmd() exec.Cmd {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Kruncmd != nil {
		return *p.Kruncmd
	}
	return exec.Cmd{}
}

func (p *AllCmds) GetGvproxyCmd() exec.Cmd {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Gvcmd != nil {
		return *p.Gvcmd
	}
	return exec.Cmd{}
}

// DO NOT BLOCK THIS FUNCTION
func WaitAPIAndPrintInfo(reortUrl string, forwardSock string, forwardState APIForwardingState, name string) error {
	if forwardState == NoForwarding {
		_ = system.KillProcess(GlobalPIDs.GetKrunkitPID())
		_ = system.KillProcess(GlobalPIDs.GetGvproxyPID())

		logrus.Errorf("Podman Rest API No forwarding")
		return nil
	}

	err := WaitAndPingAPI("unix:///" + forwardSock)
	if err != nil {
		// Why not stop krunkit and gvproxy, because I want get into machine when problem occurs, and I can debug it.
		// DO NOT STOP the krunkit and gvproxy !
		logrus.Error("failed to ping Podman API: ", err)
		return err
	} else {
		network.Reporter.SendEventToOvmJs("ready", "")
		logrus.Infof("Podman API forwarding listening on: %s\n", forwardSock)
		fmt.Printf("Podman API forwarding listening on: %s\n", forwardSock)
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

	timeout := time.After(10 * time.Second)
	res := &network.APIResponse{}
pingLoop:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout reached while waiting for Podman API")
		default:
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
