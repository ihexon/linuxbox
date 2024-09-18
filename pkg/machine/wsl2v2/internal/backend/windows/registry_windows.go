package windows

import (
	"bauklotze/pkg/machine/wsl2v2/internal/backend"
	decorate "bauklotze/pkg/machine/wsl2v2/internal/utils"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io/fs"
	"os/exec"
	"path/filepath"
	"syscall"
)

// RegistryKey wraps around a Windows registry key.
// Create it by calling OpenLxssRegistry. Must be closed after use with RegistryKey.close.
type RegistryKey struct {
	key  registry.Key
	path string // For error message purposes
}

// OpenLxssRegistry opens a registry key at the chosen path.
func (Backend) OpenLxssRegistry(path string) (r backend.RegistryKey, err error) {
	const lxssPath = `Software\Microsoft\Windows\CurrentVersion\Lxss\` // Path to the Lxss registry key. All WSL info is under this path

	p := filepath.Join(lxssPath, path)
	defer decorate.OnError(&err, "registry: could not open HKEY_CURRENT_USER\\%s", p)

	k, err := registry.OpenKey(registry.CURRENT_USER, p, registry.READ)
	if err != nil {
		return nil, err
	}

	return &RegistryKey{
		path: p,
		key:  k,
	}, nil
}

// Close releases the key.
func (r *RegistryKey) Close() (err error) {
	defer decorate.OnError(&err, "registry: could not close HKEY_CURRENT_USER\\%s", r.path)
	return r.key.Close()
}

// Field obtains the value of a Field. The value must be a string.
func (r *RegistryKey) Field(name string) (value string, err error) {
	defer decorate.OnError(&err, "registry: could not access string field %s in HKEY_CURRENT_USER\\%s", name, r.path)

	value, _, err = r.key.GetStringValue(name)
	if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
		return value, fs.ErrNotExist
	}
	if err != nil {
		return value, err
	}
	return value, nil
}

// SubkeyNames returns a slice containing the names of the current key's children.
func (r *RegistryKey) SubkeyNames() (subkeys []string, err error) {
	defer decorate.OnError(&err, "registry: could not access subkeys under HKEY_CURRENT_USER\\%s", r.path)

	keyInfo, err := r.key.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat parent registry key: %v", err)
	}
	return r.key.ReadSubKeyNames(int(keyInfo.SubKeyCount))
}

// RemoveAppxFamily uninstalls the Appx under the provided family name.
func (Backend) RemoveAppxFamily(ctx context.Context, packageFamilyName string) error {
	cmd := exec.CommandContext(ctx,
		"powershell.exe",
		"-NonInteractive",
		"-NoProfile",
		"-NoLogo",
		"-Command",
		`Get-AppxPackage | Where-Object -Property PackageFamilyName -eq "${env:PackageFamilyName}" | Remove-AppPackage`,
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PackageFamilyName=%q", packageFamilyName))

	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf("could not uninstall %q: %v. %s", packageFamilyName, err, out)
	}

	return nil
}
