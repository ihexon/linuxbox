package wsl

import (
	"fmt"
	"testing"
)

func TestWSL(t *testing.T) {

	out, err := WslExec(nil, "-l")
	if err == nil {
		fmt.Println(string(out))
		return
	}
	fmt.Println(string(out))
}
