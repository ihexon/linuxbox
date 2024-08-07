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
	exist, err := isWSLExist("ovm")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(exist)
}
