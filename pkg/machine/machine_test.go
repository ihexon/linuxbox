package machine

import (
	"bauklotze/pkg/machine/env"
	"testing"
)

func TestMachine(t *testing.T) {
	rtDir, err := env.GetRuntimeDir()
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(rtDir)

	dataDirOfVM, err := env.GetVMDataDir(0)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(dataDirOfVM)

	confDirOfVM, err := env.GetVMConfDir(0)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(confDirOfVM)
}
