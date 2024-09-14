package vmconfigs

type HyperVConfig struct {
	// ReadyVSock is the pipeline for the guest to alert the host
	// it is running
}

type WSLConfig struct {
	// Uses usermode networking
	UserModeNetworking bool
}

// Stubs
type AppleKrunkitConfig struct{}
