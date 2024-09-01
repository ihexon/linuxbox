package machine

import (
	"bauklotze/pkg/machine/machineDefine"
	"fmt"
	"strings"
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

func WaitAPIAndPrintInfo(forwardState APIForwardingState, name, forwardSock string, noInfo, rootful bool) {

	var fmtString string

	if name != machineDefine.DefaultMachineName {

	}

	if forwardState == NoForwarding {
		return
	}

	//WaitAndPingAPI(forwardSock)

	if !noInfo {
		fmt.Printf("API forwarding listening on: %s\n", forwardSock)
		if forwardState == DockerGlobal {
			fmt.Printf("Docker API clients default to this address. You do not need to set DOCKER_HOST.\n\n")
		} else {
			stillString := "still "
			switch forwardState {
			case NotInstalled:

				fmtString = `
The system helper service is not installed; the default Docker API socket
address can't be used by podman. `

			case MachineLocal:
				fmt.Printf("\nAnother process was listening on the default Docker API socket address.\n")
			case ClaimUnsupported:
				fallthrough
			default:
				stillString = ""
			}

			fmtString = `You can %sconnect Docker API clients by setting DOCKER_HOST using the
following command in your terminal session:

        %s
`
			prefix := ""
			if !strings.Contains(forwardSock, "://") {
				prefix = "unix://"
			}
			fmt.Printf(fmtString, stillString, GetEnvSetString("DOCKER_HOST", prefix+forwardSock))
		}
	}
}

func GetEnvSetString(env string, val string) string {
	return fmt.Sprintf("export %s='%s'", env, val)
}
