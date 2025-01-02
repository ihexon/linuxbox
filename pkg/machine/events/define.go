package events

// CurrentStage Stage is the stage of the lifecycle
var CurrentStage string

const (
	Init string = "init"
	Run  string = "run"
)

type InitStageName string

const (
	ExtractBootImage      InitStageName = "ExtractBootImage"
	InitUpdateConfig      InitStageName = "UpdateConfig"
	WriteSSHConnectConfig InitStageName = "WriteSSHConnectConfig"
	InitExit              InitStageName = "Exit"
)

type RunStageName string

const (
	RunUpdateConfig   RunStageName = "UpdateConfig"
	LoadMachineConfig RunStageName = "LoadMachineConfig"
	StartGvProxy      RunStageName = "StartGvProxy"
	StartVMProvider   RunStageName = "StartVMProvider"
	KillingGvProxy    RunStageName = "KillingGvProxy"
	KillingVMProvider RunStageName = "KillingKillingVMProvider"
	SyncMachineDisk   RunStageName = "SyncMachineDisk"
	MachineReady      RunStageName = "MachineReady"
	RunExit           RunStageName = "Exit"
)

type event struct {
	Stage string
	Name  string
	Value string
}

const (
	PlainTextContentType = "text/plain"
)
