package wsl2v2

import (
	"bauklotze/pkg/machine/wsl2v2/internal/backend"
	"bauklotze/pkg/machine/wsl2v2/internal/backend/windows"
	"bauklotze/pkg/machine/wsl2v2/internal/state"
	decorate "bauklotze/pkg/machine/wsl2v2/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path/filepath"
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
func (d *Distro) Name() string {
	return d.name
}

// isRegistered is the internal way of detecting whether a distro is registered or
// not. Use this one internally to avoid repeating error information.
func (d *Distro) isRegistered() (registered bool, err error) {
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
func (d *Distro) IsRegistered() (registered bool, err error) {
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

// Unregister is a wrapper around Win32's WslUnregisterDistribution.
// It irreparably destroys a distro and its filesystem.
func (d *Distro) Unregister() (err error) {
	defer decorate.OnError(&err, "could not unregister %q", d.name)

	if err := d.mustBeRegistered(); err != nil {
		return err
	}

	return d.backend.WslUnregisterDistribution(d.Name())
}

// Uninstall removes the distro's associated AppxPackage (if there is one)
// and unregisters the distro.
func (d *Distro) Uninstall(ctx context.Context) (err error) {
	defer decorate.OnError(&err, "Distro %q uninstall", d.name)

	guid, err := d.GUID()
	if err != nil {
		return err
	}

	k, err := d.backend.OpenLxssRegistry(fmt.Sprintf("{%s}", guid))
	if err != nil {
		return err
	}
	defer k.Close()

	packageFamilyName, err := k.Field("PackageFamilyName")
	if errors.Is(err, fs.ErrNotExist) {
		// Distro was imported, so there is no Appx associated
		return d.backend.WslUnregisterDistribution(d.Name())
	}
	if err != nil {
		return err
	}

	if err := d.backend.RemoveAppxFamily(ctx, packageFamilyName); err != nil {
		return err
	}

	return d.backend.WslUnregisterDistribution(d.Name())
}

func (d *Distro) mustBeRegistered() error {
	r, err := d.isRegistered()
	if err != nil {
		return err
	}
	if !r {
		return ErrNotExist
	}
	return nil
}

// Import creates a new distro from a source root filesystem.
func Import(ctx context.Context, distributionName, sourcePath, destinationPath string) (Distro, error) {
	err := os.MkdirAll(destinationPath, 0700)
	if err != nil {
		return Distro{}, fmt.Errorf("could not create destination path: %v", err)
	}

	stat, err := os.Stat(sourcePath)
	if err != nil {
		return Distro{}, fmt.Errorf("could not stat source path: %v", err)
	} else if stat.IsDir() {
		return Distro{}, errors.New("source path is a directory")
	}

	err = selectBackend(ctx).Import(ctx, distributionName, sourcePath, destinationPath)
	if err != nil {
		return Distro{}, err
	}

	return NewDistro(ctx, distributionName), nil
}

// fixPath deals with the fact that WslRegisterDistribuion is
// a bit picky with the path format.
func fixPath(relative string) (string, error) {
	abs, err := filepath.Abs(filepath.FromSlash(relative))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("file not found")
	}
	return abs, nil
}
func (d *Distro) Register(rootFsPath string) (err error) {
	defer decorate.OnError(&err, "could not register %s from rootfs in %s", d.name, rootFsPath)

	rootFsPath, err = fixPath(rootFsPath)
	if err != nil {
		return err
	}

	r, err := d.isRegistered()
	if err != nil {
		return err
	}
	if r {
		return errors.New("already registered")
	}

	return d.backend.WslRegisterDistribution(d.Name(), rootFsPath)
}

// Shutdown powers off all of WSL, including all other distros.
// Equivalent to:
//
//	wsl --shutdown
func Shutdown(ctx context.Context) error {
	return selectBackend(ctx).Shutdown()
}

// State returns the current state of the distro.
func (d *Distro) State() (s State, err error) {
	defer decorate.OnError(&err, "could not get distro %q's state", d.Name())

	registered, err := d.isRegistered()
	if err != nil {
		return s, err
	}
	if !registered {
		return state.NotRegistered, nil
	}

	return d.backend.State(d.Name())
}

// Terminate powers off the distro.
// Equivalent to:
//
//	wsl --terminate <distro>
func (d *Distro) Terminate() error {
	return d.backend.Terminate(d.Name())
}

func (d *Distro) SetAsDefault() error {
	return d.backend.SetAsDefault(d.Name())
}
