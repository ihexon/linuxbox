package wsl

import (
	"MyGoPj/wsl/internal"
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
	distroName := "Ubuntu"
	statee, err := WslState(distroName)
	if err == nil {
		t.Logf("%s Stat: %d, %s", distroName, statee, internal.Code2State(statee))
		return
	}
	t.Fatalf("Stat: %d, %s,\nError: %v", statee, internal.Code2State(statee), err)
}

func TestIsRegist(t *testing.T) {
	distroName := "Ubuntu"
	ret, err := IsRegist(distroName)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	t.Logf("%s IsRegist: %t", distroName, ret)
}
