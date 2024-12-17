//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"bauklotze/pkg/api/utils"
	"bauklotze/pkg/machine/env"
	provider2 "bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfigs"

	"github.com/Code-Hex/go-infinity-channel"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func getVM(vmName string) (*vmconfigs.MachineConfig, error) {
	providers = provider2.GetAll()
	for _, sprovider := range providers {
		dirs, err := env.GetMachineDirs(sprovider.VMType())
		if err != nil {
			return nil, fmt.Errorf("failed to get machine dirs: %w", err)
		}
		mcs, err := vmconfigs.LoadMachinesInDir(dirs)
		if err != nil {
			return nil, fmt.Errorf("failed to load machines in dir: %w", err)
		}
		if mc, exists := mcs[vmName]; exists {
			return mc, nil
		}
	}
	return nil, errors.New("can not find machine")
}

func exec(ctx context.Context, mc *vmconfigs.MachineConfig, command string, outCh *infinity.Channel[string], errCh chan string) error {
	key, err := os.ReadFile(mc.SSH.IdentityPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	connCfg := &ssh.ClientConfig{
		User:            mc.SSH.RemoteUsername,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", mc.SSH.Port), connCfg)
	if err != nil {
		errCh <- fmt.Sprintf("dial ssh error %s", err)
		return fmt.Errorf("dial ssh error %w", err)
	}
	defer conn.Close()
	context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})

	session, err := conn.NewSession()
	if err != nil {
		errCh <- fmt.Sprintf("create session error %s", err)
		return fmt.Errorf("create session error %w", err)
	}
	defer session.Close()

	w := ch2Writer(outCh)
	session.Stdout = w
	stderr := recordWriter(w)
	session.Stderr = stderr

	logrus.Infof("starting exec command: '%s'", command)
	if err := session.Run(command); err != nil {
		newErr := fmt.Errorf("%s\n%s", stderr.LastRecord(), err) //nolint:errorlint
		errCh <- newErr.Error()
		return fmt.Errorf("run command error %w", newErr)
	}
	logrus.Infof("exec command finished")

	return nil
}

type execBody struct {
	Command string `json:"command"`
}

func DoExec(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request /exec")

	name := utils.GetName(r)
	mc, err := getVM(name)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		return
	}

	var body execBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if _, ok := w.(http.Flusher); !ok {
		utils.Error(w, http.StatusInternalServerError, errors.New("streaming unsupported"))
		return
	}

	outCh := infinity.NewChannel[string]()
	errCh := make(chan string, 1)
	doneCh := make(chan struct{}, 1)

	go func() {
		if err := exec(r.Context(), mc, body.Command, outCh, errCh); err != nil {
			logrus.Warn(err.Error())
		}

		doneCh <- struct{}{}
		outCh.Close()
		close(errCh)
	}()

	defer func() {
		select {
		case <-r.Context().Done():
		default:
			_, _ = fmt.Fprintf(w, "event: done\n")
			_, _ = fmt.Fprintf(w, "data: done\n\n") // end of date
			w.(http.Flusher).Flush()
		}
	}()

	for {
		select {
		case <-doneCh:
			logrus.Infof("Command execution finished")
			return
		case err, ok := <-errCh:
			if !ok {
				continue
			}
			logrus.Warnf("Command execution error: %s", err)
			_, _ = fmt.Fprintf(w, "event: error\n")
			_, _ = fmt.Fprintf(w, "data: %s\n\n", encodeSSE(err))
			w.(http.Flusher).Flush()
			continue
		case out, ok := <-outCh.Out():
			if !ok {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: out\n")
			_, _ = fmt.Fprintf(w, "data: %s\n\n", encodeSSE(out))
			w.(http.Flusher).Flush()
			continue
		case <-r.Context().Done():
			logrus.Warnf("Client disconnected")
			return
		case <-time.After(3 * time.Second): //nolint:mnd
			_, _ = fmt.Fprintf(w, ": ping\n\n")
			w.(http.Flusher).Flush()
			continue
		}
	}
}

type chWriter struct {
	ch *infinity.Channel[string]
	mu sync.Mutex
}

func (w *chWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ch.In() <- string(p)
	return len(p), nil
}

func ch2Writer(ch *infinity.Channel[string]) io.Writer {
	return &chWriter{ch: ch}
}

type writer struct {
	w    io.Writer
	last []byte
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.last = p
	return w.w.Write(p) //nolint:wrapcheck
}

func (w *writer) LastRecord() string {
	return string(w.last)
}

func recordWriter(w io.Writer) *writer {
	return &writer{w: w}
}

func encodeSSE(str string) string {
	return strings.ReplaceAll(strings.TrimSpace(str), "\n", "\ndata: ")
}
