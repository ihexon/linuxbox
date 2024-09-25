//go:build darwin && arm64

package machine

import (
	"bauklotze/pkg/system"
	"time"
)

func WaitingAndKillProcess(ovmppid, krunkit, gvproxy int) {
	somethingWrong := make(chan bool)
	go func() {
		for {
			if ovmppid != -1 {
				if ok := system.IsProcessAlive(ovmppid); !ok {
					somethingWrong <- true
					return
				}
			}
			if err := system.CheckProcessRunning("Krunkit", krunkit); err != nil {
				somethingWrong <- true
				return
			}
			if err := system.CheckProcessRunning("GVproxy", gvproxy); err != nil {
				somethingWrong <- true
				return
			}
			// lets poll status every half second
			time.Sleep(400 * time.Millisecond)
		}
	}()

	// wait for either socket or to be ready or process to have exited
	select {
	case exited := <-somethingWrong:
		if exited == true {
			system.KillProcess(GlobalPIDs.GetGvproxyPID())
			system.KillProcess(GlobalPIDs.GetKrunkitPID())
		}
	}
}
