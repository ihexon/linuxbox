package shim

const (
	defaultGuestSock = "/run/user/%d/ovm/ovm_guest.sock"
)

// TODO: GVProxy
//func startNetworking(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider) (string, machine.APIForwardingState, error) {
//
//}

//func startHostForwarder(mc *vmconfigs.MachineConfig, provider vmconfigs.VMProvider, dirs *define.MachineDirs, hostSocks []string) error {
//	forwardUser := mc.SSH.RemoteUsername
//	guestSock := fmt.Sprintf(defaultGuestSock)
//	cfg, err := config.Default()
//	if err != nil {
//		return err
//	}
//	binary, err := cfg.FindHelperBinary(machine.ForwarderBinaryName, false)
//
//}
