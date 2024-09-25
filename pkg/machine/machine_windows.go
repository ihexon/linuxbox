//go:build windows && amd64

package machine

func WaitingAndKillProcess(ovmppid, krunkit, gvproxy int) {
	somethingWrong := make(chan bool)
	go func() {
		for {
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
