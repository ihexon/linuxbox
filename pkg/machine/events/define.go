//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package events

// CurrentStage Stage is the stage of the lifecycle

const (
	Init string = "init"
	Run  string = "run"
)

type InitStageName string

const (
	InitNewMachine   InitStageName = "InitNewMachine"
	ExtractBootImage InitStageName = "ExtractBootImage"
	InitUpdateConfig InitStageName = "UpdateConfig"
	InitSuccess      InitStageName = "Success"
	InitExit         InitStageName = "Exit"
)

type RunStageName string

const (
	LoadMachineConfig RunStageName = "LoadMachineConfig"
	StartGvProxy      RunStageName = "StartGvProxy"
	StartKrunKit      RunStageName = "StartKrunkit"
	SyncMachineDisk   RunStageName = "SyncMachineDisk"
	Ready             RunStageName = "Ready"
	RunExit           RunStageName = "Exit"
)

const (
	kError string = "error"
)
