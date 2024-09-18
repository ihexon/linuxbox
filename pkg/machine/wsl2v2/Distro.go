package wsl2v2

import (
	"bauklotze/pkg/machine/wsl2v2/internal/backend"
	"bauklotze/pkg/machine/wsl2v2/internal/backend/windows"
	"bauklotze/pkg/machine/wsl2v2/internal/state"
	decorate "bauklotze/pkg/machine/wsl2v2/internal/utils"
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

var ErrNotExist = windows.ErrNotExist

type Distro struct {
	backend backend.Backend
	name    string
}

func NewDistro(ctx context.Context, name string) Distro {
	return Distro{
		backend: selectBackend(ctx),
		name:    name,
	}
}

type State = state.State

const (
	Stopped       = state.Stopped
	Running       = state.Running
	Installing    = state.Installing
	Uninstalling  = state.Uninstalling
	NonRegistered = state.NotRegistered
)

// Name is a getter for the DistroName as shown in "wsl.exe --list".
func (d Distro) Name() string {
	return d.name
}

// isRegistered is the internal way of detecting whether a distro is registered or
// not. Use this one internally to avoid repeating error information.
func (d Distro) isRegistered() (registered bool, err error) {
	defer decorate.OnError(&err, "could not determine if distro is registered")
	distros, err := RegisteredDistros(d.backend)
	if err != nil {
		return false, err
	}

	for name := range distros {
		if strings.EqualFold(name, d.Name()) {
			return true, nil
		}
	}

	return false, nil
}

// RegisteredDistros returns a map of the registered distros and their GUID.
func RegisteredDistros(backend backend.Backend) (distros map[string]uuid.UUID, err error) {
	r, err := backend.OpenLxssRegistry(".")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	subkeys, err := r.SubkeyNames()
	if err != nil {
		return distros, err
	}

	distros = make(map[string]uuid.UUID, len(subkeys))
	for _, key := range subkeys {
		guid, err := uuid.Parse(key)
		if err != nil {
			continue // Not a WSL distro
		}

		r, err = backend.OpenLxssRegistry(key)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		name, err := r.Field("DistributionName")
		if err != nil {
			return nil, err
		}

		distros[name] = guid
	}

	return distros, nil
}

// IsRegistered returns a boolean indicating whether a distro is registered or not.
func (d Distro) IsRegistered() (registered bool, err error) {
	r, err := d.isRegistered()
	if err != nil {
		return false, fmt.Errorf("%s: %v", d.name, err)
	}
	return r, nil
}

// GUID returns the Global Unique IDentifier for the distro.
func (d *Distro) GUID() (id uuid.UUID, err error) {
	defer decorate.OnError(&err, "could not obtain GUID of %s", d.name)

	distros, err := RegisteredDistros(d.backend)
	if err != nil {
		return id, err
	}
	id, ok := distros[d.Name()]
	if !ok {
		return id, ErrNotExist
	}
	return id, nil
}
