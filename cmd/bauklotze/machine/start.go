package machine

import (
	"bauklotze/cmd/registry"
	"bauklotze/pkg/machine"
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/env"
	"bauklotze/pkg/machine/shim"
	"bauklotze/pkg/machine/vmconfigs"
	"bauklotze/pkg/network"
	"bauklotze/pkg/system"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"net/url"
	"time"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: machinePreRunE, // Get Provider and set workdir if needed
		RunE:              start,
		Args:              cobra.MaximumNArgs(1),
		Example:           `bauklotze machine start`,
	}
	startOpts = define.StartOptions{}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})
	flags := startCmd.Flags()

	twinPid := ppid
	flags.Int32Var(&startOpts.TwinPid, twinPid, -1, "the pid of PPID")

	reportUrl := reportUrl
	flags.StringVar(&startOpts.ReportUrl, reportUrl, "", "Report events to the url")
}

func start(cmd *cobra.Command, args []string) error {
	var err error
	vmName := defaultMachineName
	if len(args) > 0 && len(args[0]) > 0 {
		vmName = args[0]
	}
	dirs, err := env.GetMachineDirs(provider.VMType())
	if err != nil {
		return err
	}
	mc, err := vmconfigs.LoadMachineByName(vmName, dirs)
	if err != nil {
		return err
	}

	logrus.Infof("starting machine %q\n", vmName)
	if err = shim.Start(mc, provider, dirs, startOpts); err != nil {
		return err
	}

	if startOpts.ReportUrl != "" {
		connCtx, err := network.NewConnection(startOpts.ReportUrl)
		if err != nil {
			logrus.Errorf("Failed to connect to %q: %v\n", startOpts.ReportUrl, err)
		}
		connCtx.UrlParameter = url.Values{
			"event":   []string{"running"},
			"message": []string{"ready"},
		}
		// ? Should I return error ?
		_, _ = connCtx.DoRequest("GET", "/notify", nil)
	}

	logrus.Infof("Machine %q started successfully\n", vmName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := waiteAndStopMachine(
			ctx,
			startOpts,
			args,
			machine.GlobalPIDs.GetKrunkitPID(),
			machine.GlobalPIDs.GetGvproxyPID(),
		); err != nil {
			return err
		}
		return nil
	})

	// Start func2 in a goroutine
	g.Go(func() error {
		listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"
		if err := startRestApi(ctx, listenPath); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Println("Error:", err)
	}

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//g, ctx := errgroup.WithContext(ctx)
	//
	//g.Go(func() error {
	//	return waiteAndStopMachine(
	//		ctx,
	//		startOpts,
	//		args,
	//		machine.GlobalPIDs.GetKrunkitPID(),
	//		machine.GlobalPIDs.GetGvproxyPID(),
	//	)
	//})
	//
	//g.Go(func() error {
	//	listenPath := "unix:///" + dirs.RuntimeDir.GetPath() + "/ovm_restapi.socks"
	//	s, err := startRestApi(listenPath)
	//	if err != nil {
	//		return err
	//	}
	//
	//	context.AfterFunc(ctx, func() {
	//		s.Close()
	//	})
	//
	//	return s.Run()
	//})
	//
	//if err := g.Wait(); err != nil {
	//	return stop(nil, args)
	//}
	//
	//return err
	return err
}

func startRestApi(ctx context.Context, listenPath string) error {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		if err := service(nil, []string{listenPath}); err != nil {
			cancel(err)
			//errChan <- err
		}
	}()

	<-ctx.Done()
	return context.Cause(ctx)
}

func waiteAndStopMachine(ctx context.Context, startOpts define.StartOptions, args []string, krunkit, gvproxy int) error {
	ctx, cancel := context.WithCancelCause(ctx)

	var err error
	// If user do not --twinpid, get my PPID
	if startOpts.TwinPid == -1 {
		startOpts.TwinPid, err = system.GetMyPPID()
		if err != nil {
			return err
		}
	}
	logrus.Infof("Waiting PPID[%d] exited then stop the machine\n", startOpts.TwinPid)

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
			if !system.IsProcessAlive(int(startOpts.TwinPid)) {
				cancel(fmt.Errorf("%s exited, stop the krunkit and gvproxy", startOpts.TwinPid))
			}

			if err := system.CheckProcessRunning("KRunkit", krunkit); err != nil {
				cancel(err)
			}

			if err := system.CheckProcessRunning("GVProxy", gvproxy); err != nil {
				cancel(err)
			}
			time.Sleep(400 * time.Millisecond)
		}
	}
}
