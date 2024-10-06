package types

type APIContextKey int

const (
	DecoderKey APIContextKey = iota
	RuntimeKey
	IdleTrackerKey
	ConnKey
	CompatDecoderKey
	DefaultCORSAllowedHost = "http://127.0.0.1"
)
