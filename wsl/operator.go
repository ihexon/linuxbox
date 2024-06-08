package wsl

import (
	"MyGoPj/shell"
	"context"
	"fmt"
)

// TODO:  WSL2 Shutdown()
func Shutdown() error {
	return nil
}

func WslExec(ctx context.Context, args ...string) ([]byte, error) {

	envs := []string{"WSL_UTF8=1"}
	out, err := shell.Exec(ctx, envs, "wsls.exe", args...)
	if err == nil {
		return out, nil
	}
	fmt.Println(err.Error())
	return out, err
}
