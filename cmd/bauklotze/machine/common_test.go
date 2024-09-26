package machine

import (
	"bauklotze/pkg/utils/reexec"
	"testing"
)

func TestMachine(t *testing.T) {
	s := reexec.Command("fuck", "me", "plz")
	t.Log(s)
}
