package machine

type APIForwardingState int

var (
	ForwarderBinaryName = "gvproxy"
)

const (
	NoForwarding APIForwardingState = iota
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
