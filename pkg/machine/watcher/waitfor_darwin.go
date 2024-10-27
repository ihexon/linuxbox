//go:build darwin && (arm64 || amd64)

package watcher

import (
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/system"
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/url"
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

// WaitApiServerAndStopMachine is Non block function
func WaitApiServerAndStopMachine(g *errgroup.Group, ctx context.Context, dirs *define.MachineDirs) {
	listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"
	logrus.Infof("Starting API Server in %s\n", listenPath)

	g.Go(func() error {
		apiURL, _ := url.Parse(listenPath)
		return startRestApi(ctx, apiURL)
	})
}

func startRestApi(ctx context.Context, apiURL *url.URL) error {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		if err := server.RestService(apiURL); err != nil {
			cancel(err)
		}
	}()
	<-ctx.Done()
	return context.Cause(ctx)
}
