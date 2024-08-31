package machine

type APIForwardingState int

type StartOptions struct {
	NoInfo  bool
	Quiet   bool
	Rosetta bool
}

type StopOptions struct{}

type RemoveOptions struct {
	Force        bool
	SaveImage    bool
	SaveIgnition bool
}
type ResetOptions struct {
	Force bool
}
