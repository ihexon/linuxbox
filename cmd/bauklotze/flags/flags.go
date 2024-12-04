//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package cmdflags

const (
	WorkspaceFlag   = "workspace"
	ReportUrlFlag   = "report-url"
	BootImageFlag   = "boot"
	BootVersionFlag = "boot-version"
	DataVersionFlag = "data-version"
	VolumeFlag      = "volume"
	MemoryFlag      = "memory"
	CpusFlag        = "cpus"
	PpidFlag        = "ppid"
	LogOutFlag      = "log-out"
	LogLevelFlag    = "log-level"
)

const (
	// 	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	DefaultLogLevel    = "info"
	ConsoleBased       = "console"
	FileBased          = "file"
	MaxMachineNameSize = 30
	KrunMaxCpus        = 8
)

const (
	BAUKLOTZE_HOME = "BAUKLOTZE_HOME"
	TMP_DIR        = "/tmp/"
)
