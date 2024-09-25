//go:build windows

package config

func getDefaultMachineVolumes() []string {
	return []string{}
}

var defaultHelperBinariesDir = []string{
	"C:\\Program Files\\Oomol\\ovm\\bin",
}
