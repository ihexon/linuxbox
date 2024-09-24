package machine

type APIForwardingState int

type StopOptions struct {
}

var (
	ForwarderBinaryName = "gvproxy"
)

const (
	NoForwarding APIForwardingState = iota
	ClaimUnsupported
	NotInstalled
	MachineLocal
	DockerGlobal
	InForwarding
)

type RemoveOptions struct {
	Force        bool
	SaveImage    bool
	SaveIgnition bool
}
type ResetOptions struct {
	Force bool
}
