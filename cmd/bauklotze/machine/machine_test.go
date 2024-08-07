package machine

import (
	provider2 "bauklotze/pkg/machine/provider"
	"testing"
)

func TestGet(t *testing.T) {
	vmProvider, err := provider2.Get()
	if err == nil {
		t.Logf("%+v\n", vmProvider)
	}
	t.Logf("%+v\n", vmProvider)
}
