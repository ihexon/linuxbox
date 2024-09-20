package env

import (
	"path/filepath"
	"testing"
)

func TestDirs(t *testing.T) {
	abs, err := filepath.Abs("c://s/ssd/asd")
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Log(abs)
}
