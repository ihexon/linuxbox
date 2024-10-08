package vhdx

import (
	"bauklotze/pkg/machine/define"
	"testing"
)

func TestVHDX(t *testing.T) {

	file, err := define.NewMachineFile("C:\\Users\\localuser\\test_block.vhdx", nil)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = VhdxCreate(file, 1073741824)
	if err != nil {
		t.Errorf(err.Error())
	}
}
