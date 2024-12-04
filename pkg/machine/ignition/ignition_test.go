//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"encoding/json"
	"net/url"
	"os"
	"testing"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
)

func TestGetDirs(t *testing.T) {
	ignBuilder := NewIgnitionBuilder(DynamicIgnitionV2{
		Name:     DefaultIgnitionUserName,
		Key:      "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone: "local", // Auto detect timezone from locales
		VMType:   define.LibKrun,
		VMName:   define.DefaultMachineName,
		MachineConfigs: &vmconfigs.MachineConfig{
			Mounts: []*vmconfigs.Mount{
				{
					Type:   "virtiofs",
					Tag:    "virtio-zzh",
					Source: "/zzh",
					Target: "/mnt/zzh",
				}, {
					Type:   "virtiofs",
					Tag:    "virtio-zzh1",
					Source: "/zzh1",
					Target: "/mnt/zzh1",
				},
			},
		},
		WritePath: "/tmp/generateConfig.json",
		Rootful:   true,
	})

	dirs := ignBuilder.dynamicIgnition.getDirs("root")
	jsonDirs, err := json.MarshalIndent(dirs, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
}

func TestGetUser(t *testing.T) {
	ignBuilder := NewIgnitionBuilder(DynamicIgnitionV2{
		Name:      "root",
		Key:       "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone:  "myTimeZone",
		VMType:    define.LibKrun,
		VMName:    "VMName",
		WritePath: "/tmp/generateConfig.json",
		Rootful:   true,
	})

	user := ignBuilder.dynamicIgnition.getUsers()
	jsonDirs, err := json.MarshalIndent(user, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
}

func TestGetFiles(t *testing.T) {
	ignBuilder := NewIgnitionBuilder(DynamicIgnitionV2{
		Name:     DefaultIgnitionUserName,
		Key:      "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone: "local", // Auto detect timezone from locales
		VMType:   define.LibKrun,
		VMName:   define.DefaultMachineName,
		MachineConfigs: &vmconfigs.MachineConfig{
			Mounts: []*vmconfigs.Mount{
				{
					Type:   "virtiofs",
					Tag:    "virtio-zzh",
					Source: "/zzh",
					Target: "/mnt/zzh",
				}, {
					Type:   "virtiofs",
					Tag:    "virtio-zzh1",
					Source: "/zzh1",
					Target: "/mnt/zzh1",
				},
			},
		},
		WritePath: "/tmp/generateConfig.json",
		Rootful:   true,
	})

	files := ignBuilder.dynamicIgnition.getFiles("root", 0, define.LibKrun)
	jsonDirs, err := json.MarshalIndent(files, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
}

func TestGetLinks(t *testing.T) {
	ignBuilder := NewIgnitionBuilder(DynamicIgnitionV2{
		Name:     DefaultIgnitionUserName,
		Key:      "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone: "local", // Auto detect timezone from locales
		VMType:   define.LibKrun,
		VMName:   define.DefaultMachineName,
		MachineConfigs: &vmconfigs.MachineConfig{
			Mounts: []*vmconfigs.Mount{
				{
					Type:   "virtiofs",
					Tag:    "virtio-zzh",
					Source: "/zzh",
					Target: "/mnt/zzh",
				}, {
					Type:   "virtiofs",
					Tag:    "virtio-zzh1",
					Source: "/zzh1",
					Target: "/mnt/zzh1",
				},
			},
		},
		WritePath: "/tmp/generateConfig.json",
		Rootful:   true,
	})

	links := ignBuilder.dynamicIgnition.getLinks("root")
	jsonDirs, err := json.MarshalIndent(links, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
}

func TestDynamicIgnitionV2_GenerateIgnitionConfig(t *testing.T) {
	ignBuilder := NewIgnitionBuilder(DynamicIgnitionV2{
		Name:     DefaultIgnitionUserName,
		Key:      "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone: "local", // Auto detect timezone from locales
		VMType:   define.LibKrun,
		VMName:   define.DefaultMachineName,
		MachineConfigs: &vmconfigs.MachineConfig{
			Mounts: []*vmconfigs.Mount{
				{
					Type:   "virtiofs",
					Tag:    "virtio-zzh",
					Source: "/zzh",
					Target: "/mnt/zzh",
				}, {
					Type:   "virtiofs",
					Tag:    "virtio-zzh1",
					Source: "/zzh1",
					Target: "/mnt/zzh1",
				},
			},
		},
		WritePath: "/tmp/generateConfig.json",
		Rootful:   true,
	})

	err := ignBuilder.dynamicIgnition.GenerateIgnitionConfig()
	if err != nil {
		t.Fatalf("Failed to generate ignition config: %v", err)
	}

	cfg := ignBuilder.dynamicIgnition.Cfg
	jsonDirs, err := json.MarshalIndent(cfg, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
	err = ignBuilder.Build()
	if err != nil {
		t.Error(err)
	}
}

func TestIgnServer(t *testing.T) {
	addr := "tcp://127.0.0.1:8899"
	listener, err := url.Parse(addr)
	if err != nil {
		t.Error(err.Error())
	}
	fileStr := "C:\\Users\\localuser\\Bauklotze\\README.md"

	file, err := os.Open(fileStr)
	if err != nil {
		t.Error(err.Error())
	}
	errChan := make(chan error, 1)
	err = ServeIgnitionOverSocketCommon(listener, file)
	if err != nil {
		errChan <- err
	}

	err = <-errChan
	t.Log(err.Error())
}
