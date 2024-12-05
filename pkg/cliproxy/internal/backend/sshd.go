//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gliderlabs/ssh"
)

func SSHD() error {
	ssh.Handle(func(s ssh.Session) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			<-s.Context().Done()
			cancel()
		}()

		str := s.Command()
		if len(str) == 0 {
			return
		}

		_, _ = fmt.Fprintf(os.Stdout, "Proxy command: %s\n", str)
		cmd := exec.CommandContext(ctx, str[0], str[1:]...)
		stdOut, handleErr := cmd.StdoutPipe()
		if handleErr != nil {
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", handleErr)
			return
		}

		stdErr, handleErr := cmd.StderrPipe()
		if handleErr != nil {
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", handleErr)
			return
		}

		handleErr = cmd.Start()
		if handleErr != nil {
			_ = s.Exit(127)
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", handleErr)
		}

		go func() {
			_, _ = io.Copy(s, stdOut)
		}()
		go func() {
			_, _ = io.Copy(s.Stderr(), stdErr)
		}()

		if err := cmd.Wait(); err != nil {
			_ = s.Exit(cmd.ProcessState.ExitCode())
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Command: %s finished\n", str)
		}
	})

	if err := ssh.ListenAndServe("127.0.0.1:5321", nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return fmt.Errorf("listen and serve error: %w", err)
	}

	return nil
}
