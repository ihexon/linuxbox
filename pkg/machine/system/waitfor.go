package system

import (
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/machine/define"
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/url"
	"time"
)

func WaitProcessAndStopMachine(g *errgroup.Group, ctx context.Context, ovmPid, krunPid, gvPid int32) {
	g.Go(func() error {
		return watchProcess(ctx, ovmPid, krunPid, gvPid)
	})
}

func watchProcess(ctx context.Context, ovmPid, krunPid, gvPid int32) error {
	logrus.Infof("Waiting PPID[%d], krunkit[%d],gvproxy[%d] exited then stop the machine\n", ovmPid, krunPid, gvPid)
	pids := []int32{(ovmPid), (krunPid), (gvPid)}
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if isRunning, err := IsProcesSAlive(pids); !isRunning {
			return err
		}
		time.Sleep(400 * time.Millisecond)
	}
	return nil
}

func WaitApiServerAndStopMachine(g *errgroup.Group, ctx context.Context, dirs *define.MachineDirs) {

	g.Go(func() error {
		listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"
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
