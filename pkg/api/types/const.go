package types

type APIContextKey int

const (
	DecoderKey APIContextKey = iota
	ConnKey
	DefaultCORSAllowedHost = "http://127.0.0.1"
)
