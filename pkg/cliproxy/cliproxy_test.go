package cliproxy

import (
	"os"
	"testing"
)

func TestCliProxy(t *testing.T) {
	_ = os.Setenv("BAUKLOTZE_HOME", "/tmp/_ovm_test_temp_dir")
	_ = os.MkdirAll(os.Getenv("BAUKLOTZE_HOME"), 0755)
	err := RunCliProxy()

	if err != nil {
		t.Errorf(err.Error())
	}
}
