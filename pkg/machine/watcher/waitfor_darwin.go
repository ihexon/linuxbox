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
func WaitProcessAndStopMachine(g *errgroup.Group, ctx context.Context, ovmPid, krunPid, gvPid int32) {
	g.Go(func() error {
		return watchProcess(ctx, ovmPid, krunPid, gvPid)
	})
}

func watchProcess(ctx context.Context, ovmPid, krunPid, gvPid int32) error {
	logrus.Infof("Waiting PPID[%d], krunkit[%d], gvproxy[%d] exited then stop the machine\n", ovmPid, krunPid, gvPid)
	pids := []int32{(ovmPid), (krunPid), (gvPid)}
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if isRunning, err := system.IsProcesSAlive(pids); !isRunning {
			return err
		}
		time.Sleep(400 * time.Millisecond)
	}
	return nil
}

// WaitApiServerAndStopMachine is Non block function
func WaitApiServerAndStopMachine(g *errgroup.Group, ctx context.Context, dirs *define.MachineDirs) {
	listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"

	g.Go(func() error {
		apiURL, _ := url.Parse(listenPath)
		return startRestApi(ctx, apiURL)
	})

	logrus.Infof("Starting API Server in %s\n", listenPath)
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
