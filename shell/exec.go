package shell

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
)

func Exec(ctx context.Context, envs []string, prog string, args ...string) ([]byte, error) {

	var out bytes.Buffer

	cmd := exec.Command(prog, args...)

	for i, e := range envs {
		_, _, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		cmd.Env = append(os.Environ(), envs[i])
	}

	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		return out.Bytes(), nil
	}
	return out.Bytes(), err
}
