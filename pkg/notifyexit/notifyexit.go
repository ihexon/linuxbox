package notifyexit

import (
	"bauklotze/pkg/network"
	"os"
)

func NotifyExit(code int) {
	network.Reporter.SendEventToOvmJs("exit", "")
	os.Exit(code)
}

func JustExit(code int) {
	os.Exit(code)
}
