//go:build darwin && (arm64 || amd64)

package watcher

import (
	"bauklotze/pkg/machine/system"
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"time"
)

// WaitProcessAndStopMachine is Non block function
func WaitProcessAndStopMachine(g *errgroup.Group, ctx context.Context, ovmPPid, krunPid, gvPid int32) {
	g.Go(func() error {
		return watchProcess(ctx, ovmPPid, krunPid, gvPid)
	})
	logrus.Infof("Waiting for PPID [ %d ] Krun [ %d ] and GV [ %d ] to stop\n", ovmPPid, krunPid, gvPid)
}

func watchProcess(ctx context.Context, ovmPPid, krunPid, gvPid int32) error {
	pids := []int32{(ovmPPid), (krunPid), (gvPid)}
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if isRunning, err := system.IsProcesSAlive(pids); !isRunning {
			return err
		}
		time.Sleep(300 * time.Millisecond)
	}
}
