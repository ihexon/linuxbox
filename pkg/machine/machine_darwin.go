//go:build darwin && arm64

package machine

import (
	"bauklotze/pkg/system"
	"context"
	"fmt"
	"time"
)

// Function TwinPidKiller checks if the given PID is alive and wait.
// Kills gvproxy and krunkit if the PID is not alive
func TwinPidKiller(pid_wait, pid_k1, pid_k2 int) {
	pidExited := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer close(pidExited)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println("Continue checking...")
			}
			if ok := system.IsProcessAlive(pid_wait); !ok {
				pidExited <- true
				return
			}
			// lets poll status every half second
			time.Sleep(400 * time.Millisecond)
		}
	}()

	// wait for either socket or to be ready or process to have exited
	select {
	case exited := <-pidExited:
		if exited == true {
			system.KillProcess(pid_k1)
			system.KillProcess(pid_k2)
		}
	}
}
