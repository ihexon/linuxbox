package env

import (
	"bauklotze/pkg/machine/provider"
	"bauklotze/pkg/machine/vmconfig"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSSHIdentityPath(t *testing.T) {
	vmp, err := getVMProvider()
	if err != nil {
		t.Errorf("getVMProvider() failed: %v", err)
	}

	osTempDir := os.TempDir()
	if osTempDir == "" || osTempDir == "/" {
		t.Fatalf("os.TempDir() failed: %v", osTempDir)
	}

	err = setWorkSpace(filepath.Join(osTempDir, "test_workspace"))
	if err != nil {
		t.Errorf("setWorkSpace() failed: %v", err)
	}
	defer os.RemoveAll(filepath.Join(osTempDir, "test_workspace"))

	path, err := vmconfig.GetSSHIdentityPath(vmp.VMType())
	if err != nil {
		t.Errorf("GetSSHIdentityPath() failed: %v", err)
	}
	t.Log(path.GetPath())
}

func TestGetVMProvider(t *testing.T) {
	vmp, err := getVMProvider()
	if err != nil {
		t.Errorf("getVMProvider() failed: %v", err)
	}
	t.Logf("VMProvider: %s", vmp.VMType().String())
}

func TestWorkspace(t *testing.T) {
	osTempDir := os.TempDir()
	if osTempDir == "" || osTempDir == "/" {
		t.Fatalf("os.TempDir() failed: %v", osTempDir)
	}

	err := setWorkSpace(filepath.Join(osTempDir, "test_workspace"))
	if err != nil {
		t.Errorf("setWorkSpace() failed: %v", err)
	}
	defer os.RemoveAll(filepath.Join(osTempDir, "test_workspace"))

	ws, err := vmconfig.GetWorkSpace()
	if err != nil {
		t.Errorf("GetWorkSpace() failed: %v", err)
	}
	t.Logf("Workspace: %s", ws.GetPath())
}

func TestGetMachineDirs(t *testing.T) {
	vmp, err := getVMProvider()
	if err != nil {
		t.Errorf("getVMProvider() failed: %v", err)
	}

	osTempDir := os.TempDir()
	if osTempDir == "" || osTempDir == "/" {
		t.Fatalf("os.TempDir() failed: %v", osTempDir)
	}
	err = setWorkSpace(filepath.Join(osTempDir, "test_workspace"))
	if err != nil {
		t.Errorf("setWorkSpace() failed: %v", err)
	}
	defer os.RemoveAll(filepath.Join(osTempDir, "test_workspace"))

	dirs, err := vmconfig.GetMachineDirs(vmp.VMType())
	if err != nil {
		t.Errorf("GetMachineDirs() failed: %v", err)
	}
	t.Logf("DataDir: %s", dirs.DataDir.Path)
	t.Logf("TmpDir: %s", dirs.TmpDir.Path)
	t.Logf("LogDir: %s", dirs.LogsDir.Path)
	t.Logf("ConfigDir: %s", dirs.ConfigDir.Path)
}

func setWorkSpace(p string) error {
	_, err := vmconfig.SetWorkSpace(p)
	if err != nil {
		return fmt.Errorf("SetWorkSpace() failed: %w", err)
	}
	return nil
}

func getVMProvider() (vmconfig.VMProvider, error) {
	vmp, err := provider.Get()
	if err != nil {
		return nil, fmt.Errorf("provider.Get() failed: %w", err)
	}
	return vmp, nil
}
