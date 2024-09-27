//go:build darwin && arm64

package config

func getDefaultMachineVolumes() []string {
	return []string{
		"/Users:/mnt/fromHost/Users",
		"/private:/mnt/fromHost/private",
		"/var/folders:/mnt/fromHost/var/folders",
	}
}

var defaultHelperBinariesDir = []string{
	// Relative to the binary directory
	"$BINDIR/../libexec/podman",
	// Homebrew install paths
	"/usr/local/opt/podman/libexec/podman",
	"/opt/homebrew/opt/podman/libexec/podman",
	"/opt/homebrew/bin",
	"/usr/local/bin",
	// default paths
	"/usr/local/libexec/podman",
	"/usr/local/lib/podman",
	"/usr/libexec/podman",
	"/usr/lib/podman",
}
