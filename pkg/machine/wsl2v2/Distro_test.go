package wsl2v2

import (
	"context"
	"testing"
)

func Test_registeredDistros(t *testing.T) {
	distros, err := RegisteredDistros(selectBackend(context.Background()))
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	t.Log(distros)
}
