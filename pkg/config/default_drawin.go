//go:build darwin && arm64

package config

func getDefaultMachineVolumes() []string {
	// Empty mount point
	return []string{}
}

var defaultHelperBinariesDir = []string{
	// Relative to the binary directory
	"$BINDIR/../libexec/",
}
