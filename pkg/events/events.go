package events

type Status string

const (
	Init  Status = "init"
	Kill  Status = "kill"
	Start Status = "start"
	Stop  Status = "stop"
)

type Event struct {
	Name string `json:",omitempty"`
}
