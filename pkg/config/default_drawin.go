package config

func getDefaultMachineVolumes() []string {
	return []string{
		"/Users:/Users",
		"/private:/private",
		"/var/folders:/var/folders",
	}
}
