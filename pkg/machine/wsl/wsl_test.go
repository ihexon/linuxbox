//go:build !drawin && !linux && windows

package wsl

import (
	"testing"
)

func TestWSLFuncs(t *testing.T) {
	distros, err := getAllWSLDistros(false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for k, _ := range distros {
		t.Logf(k)
	}
}

func TestWSLStubber_Exists(t *testing.T) {
	exist, err := isWSLExist("Ubuntu")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(exist)
}

func TestWslPipe(t *testing.T) {
	wslPipe("ls /", "Ubuntu", "--help")

}
