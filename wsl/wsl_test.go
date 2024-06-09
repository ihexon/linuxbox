package wsl

import (
	"fmt"
	"testing"
)

func TestWslExec(t *testing.T) {
	out, err := WslExec(nil, "-l")
	if err == nil {
		fmt.Println(string(out))
		return
	}
	t.Fatalf("%v, %v", out, err)
}

func TestWslState(t *testing.T) {
	out, err := WslState("Ubuntu")
	if err == nil {
		t.Logf("Stat: %d", out)
		return
	}
	t.Fatalf("Stat: %d,\nError: %v", out, err)
}
