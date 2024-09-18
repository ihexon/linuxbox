package wsl2v2

import (
	"bauklotze/pkg/machine/wsl2v2/internal/state"
	"context"
)

func wslCheckExists(dist string, running bool) (bool, error) {
	distro := NewDistro(context.Background(), dist)
	s, err := distro.backend.State(distro.name)
	if err != nil {
		return false, err
	}

	switch s {
	case state.Running, state.Stopped:
		return true, nil
	default:
	}
	return false, err
}

func terminateDist(dist string) error {
	distro := NewDistro(context.Background(), dist)
	err := distro.backend.Terminate(distro.name)
	if err != nil {
		return err
	}
	return nil
}

func unregisterDist(dist string) error {
	distro := NewDistro(context.Background(), dist)
	err := distro.backend.WslUnregisterDistribution(distro.name)
	if err != nil {
		return err
	}
	return err
}

// Deprecated, Using isRunningV2
func isRunningv1(dist string) (bool, error) {
	distro := NewDistro(context.Background(), dist)
	s, err := distro.backend.State(distro.name)
	if err != nil {
		return false, err
	}
	switch s {
	case state.Running:
		return true, nil
	default:
	}
	return false, err
}

func isRunningV2(distroName string) (bool, error) {
	distro := Distro{
		backend: selectBackend(context.Background()),
		name:    distroName,
	}
	s, err := distro.State()
	if err != nil {
		return false, err
	}
	if s == state.Running {
		return true, nil

	}
	return false, err
}
