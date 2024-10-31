package backend

import (
	"context"
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"os"
	"os/exec"
)

func SSHD() error {
	var err error
	ssh.Handle(func(s ssh.Session) {
		var handleErr error
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			<-s.Context().Done()
			cancel()
		}()

		str := s.Command()
		if str == nil || len(str) == 0 {
			return
		}

		//if str[0] != "ffmpeg" && str[0] != "llama" {
		//	_, _ = fmt.Fprintf(s.Stderr(), "Command not allowed\n")
		//	return
		//}

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
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", handleErr)
		}

		go func() {
			_, _ = io.Copy(s, stdOut)
		}()
		go func() {
			_, _ = io.Copy(s.Stderr(), stdErr)
		}()

		if err := cmd.Wait(); err != nil {
			_, _ = fmt.Fprintf(s.Stderr(), "Error: %s\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Command: %s finished\n", str)
		}
		return
	})

	err = ssh.ListenAndServe("127.0.0.1:5321", nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return err
	}

	return err
}
