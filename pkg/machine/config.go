package machine

type APIForwardingState int

type StartOptions struct {
	NoInfo  bool
	Quiet   bool
	Rosetta bool
	SendEvt string
	// TwinPid is the parent process that calls Bauklotze.
	// Bauklotze will automatically exit when the parent process ends.
	TwinPid int
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
)

type StopOptions struct{}

type RemoveOptions struct {
	Force        bool
	SaveImage    bool
	SaveIgnition bool
}
type ResetOptions struct {
	Force bool
}
