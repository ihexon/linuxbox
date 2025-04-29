//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/vmconfig"
	"bauklotze/pkg/machine/volumes"

	"github.com/sirupsen/logrus"
)

type DynamicIgnitionV3 struct {
	IgnFile         *io.PathWrapper
	SSHIdentityPath *io.PathWrapper
	VMType          vmconfig.VMType
	Mounts          []volumes.Mount
	TimeZone        string
	CodeBuffer      *bytes.Buffer
}

func (ign *DynamicIgnitionV3) Write() error {
	err := ign.IgnFile.Delete(false)
	if err != nil {
		return fmt.Errorf("failed to delete ignition file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(ign.IgnFile.GetPath()), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories for ignition file: %w", err)
	}

	file, err := os.Create(ign.IgnFile.GetPath())
	if err != nil {
		return fmt.Errorf("failed to create ignition file: %w", err)
	}
	defer file.Close() //nolint:errcheck
	_, err = file.Write(ign.CodeBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write ignition file: %w", err)
	}
	logrus.Infof("Ignition file written to %s", ign.IgnFile.GetPath())
	return nil
}

func (ign *DynamicIgnitionV3) GenerateIgnitionConfig(mycode []string) error {
	ign.CodeBuffer = new(bytes.Buffer)

	err := ign.GenerateMountScripts()
	if err != nil {
		return fmt.Errorf("failed to generate mount scripts: %w", err)
	}

	err = ign.GenerateUserProvidedScripts(mycode)
	if err != nil {
		return fmt.Errorf("failed to generate user provided scripts: %w", err)
	}

	if ign.SSHIdentityPath.GetPath() != "" {
		err = ign.CopySSHIdPub()
		if err != nil {
			return fmt.Errorf("failed to copy ssh id pub: %w", err)
		}
	}

	if err = ign.UpdateTimeZone(); err != nil {
		return fmt.Errorf("failed to update timezone: %w", err)
	}

	if err = ign.GeneratePodmanMachineConfig(); err != nil {
		return fmt.Errorf("failed to generate podman machine config: %w", err)
	}

	return nil
}

// GenerateUserProvidedScripts Write the user provided scripts to the DynamicIgnitionV3.CodeBuffer
func (ign *DynamicIgnitionV3) GenerateUserProvidedScripts(mycode []string) error {
	for _, code := range mycode {
		fmt.Fprintln(ign.CodeBuffer, "# User provided script")
		fmt.Fprintf(ign.CodeBuffer, "%s\n", code)
	}
	return nil
}

func (ign *DynamicIgnitionV3) GeneratePodmanMachineConfig() error {
	t := template.Must(template.New("PodmanMachineConfigScriptCodes").Parse(podmanMachineConfigScript))
	mybuff := new(bytes.Buffer)
	data := struct {
		CurrentVMType string
	}{
		CurrentVMType: ign.VMType.String(),
	}

	if err := t.Execute(mybuff, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	ign.CodeBuffer.Write(mybuff.Bytes())
	return nil
}

// GenerateMountScripts a template for the virtiofs mount script
func (ign *DynamicIgnitionV3) GenerateMountScripts() error {
	t := template.Must(template.New("VirtioFsMountScriptCodes").Parse(VirtioFSMountScript))
	mybuff := new(bytes.Buffer)
	for _, vol := range ign.Mounts {
		if vol.Type == volumes.VirtIOFS.String() && !strings.HasPrefix(vol.Target, filepath.Dir(ign.IgnFile.GetPath())) {
			data := struct {
				FsType string
				Source string
				Target string
				Tag    string
			}{
				FsType: vol.Type,
				Source: vol.Source,
				Target: vol.Target,
				Tag:    vol.Tag,
			}
			// Execute will append the data into mybuff
			if err := t.Execute(mybuff, data); err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}
		}
	}
	ign.CodeBuffer.Write(mybuff.Bytes())
	return nil
}

func (ign *DynamicIgnitionV3) CopySSHIdPub() error {
	sshkeyData, err := os.ReadFile(ign.SSHIdentityPath.GetPath() + ".pub")
	if err != nil {
		return fmt.Errorf("failed to read ssh key: %w", err)
	}

	data := struct {
		Target string
	}{
		Target: strings.TrimSpace(string(sshkeyData)),
	}

	mybuff := new(bytes.Buffer)
	t := template.Must(template.New("WriteSSHPubKeyScriptCodes").Parse(WriteSSHPubKeyScript))
	if err := t.Execute(mybuff, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	ign.CodeBuffer.Write(mybuff.Bytes())
	return nil
}

func (ign *DynamicIgnitionV3) UpdateTimeZone() error {
	tz, err := getLocalTimeZone()
	if err != nil {
		return fmt.Errorf("failed to get local timezone: %w", err)
	}
	data := struct {
		TimeZone string
	}{
		TimeZone: tz,
	}

	mybuff := new(bytes.Buffer)
	t := template.Must(template.New("UpdateTimeZoneScriptCodes").Parse(UpdateTimeZoneScript))
	if err := t.Execute(mybuff, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	ign.CodeBuffer.Write(mybuff.Bytes())
	return nil
}

func NewIgnitionBuilder(dynamicIgnition *DynamicIgnitionV3) *DynamicIgnitionV3 {
	return dynamicIgnition
}

func getLocalTimeZone() (string, error) {
	tzPath, err := os.Readlink("/etc/localtime")
	if err != nil {
		return "", fmt.Errorf("failed to read link: %w", err)
	}
	return strings.TrimPrefix(tzPath, "/var/db/timezone/zoneinfo/"), nil
}
