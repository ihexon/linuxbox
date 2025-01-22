//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package events

// CurrentStage Stage is the stage of the lifecycle
var CurrentStage string

const (
	Init string = "init"
	Run  string = "run"
)

type InitStageName string

const (
	InitNewMachine   InitStageName = "InitNewMachine"
	ExtractBootImage InitStageName = "ExtractBootImage"
	InitUpdateConfig InitStageName = "UpdateConfig"
	InitExit         InitStageName = "Exit"
)

type RunStageName string

const (
	LoadMachineConfig RunStageName = "LoadMachineConfig"
	StartGvProxy      RunStageName = "StartGvProxy"
	StartVMProvider   RunStageName = "StartVMProvider"
	SyncMachineDisk   RunStageName = "SyncMachineDisk"
	Ready             RunStageName = "Ready"
	RunExit           RunStageName = "Exit"
)

const (
	kError string = "error"
)

type event struct {
	Stage string
	Name  string
	Value string
}

const (
	PlainTextContentType = "text/plain"
)
