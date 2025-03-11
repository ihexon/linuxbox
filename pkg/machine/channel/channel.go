package channel

import "context"

var (
	vmReady  context.Context
	vmCancel context.CancelFunc
)

func init() {
	vmReady, vmCancel = context.WithCancel(context.Background()) //nolint:fatcontext
}

func NotifyMachineReady() {
	vmCancel()
}

func WaitVMReady() <-chan struct{} {
	return vmReady.Done()
}

func IsVMReady() bool {
	select {
	case <-vmReady.Done():
		return true
	default:
		return false
	}
}

func Close() {
	vmCancel()
}
