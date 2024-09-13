package network

import (
	"bauklotze/pkg/events"
	"testing"
)

func TestSocks(t *testing.T) {
	SendEventToOvmJs(events.Start, "KunkitStaring...", "/tmp/fuck.sock")
}
