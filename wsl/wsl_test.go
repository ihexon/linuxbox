package wsl

import (
	"MyGoPj/wsl/internal"
	"context"
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
	distroName := "Ubuntus"
	ret, err := IsRegist(context.Background(), distroName)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	t.Logf("%s IsRegist: %t", distroName, ret)
}

func TestImportDistro(t *testing.T) {
	err := ImportDistro(context.Background(), true, "alpine1", "C:\\Users\\localuser\\Documents", "C:\\Users\\localuser\\Documents\\alpine-minirootfs-3.20.0-x86_64.tar.gz")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
}
