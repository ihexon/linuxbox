package machine

import (
	"bauklotze/pkg/machine/machineDefine"
	"fmt"
)

func PrintRootlessWarning(name string) {
	suffix := ""
	if name != machineDefine.DefaultMachineName {
		suffix = " " + name
	}

	fmtString := `
This machine is currently configured in rootless mode. If your containers
require root permissions (e.g. ports < 1024), or if you run into compatibility
issues with non-podman clients, you can switch using the following command:

	podman machine set --rootful%s

`
	fmt.Printf(fmtString, suffix)
}
