package ignition

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"encoding/json"
	"net/url"
	"os"
	"testing"
)

func TestGetDirs(t *testing.T) {
	dirs := getDirs("root")
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
	files := getFiles("root", 0, true, define.LibKrun, true)
	jsonDirs, err := json.MarshalIndent(files, " ", "   ")
	if err != nil {
		t.Fatalf("Failed to marshal dirs: %v", err)
	}
	t.Log(string(jsonDirs))
}

func TestGetLinks(t *testing.T) {
	links := getLinks("root")
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
}

func TestIgnServer(t *testing.T) {
	addr := "tcp://127.0.0.1:8899"
	listener, err := url.Parse(addr)
	if err != nil {
		t.Errorf(err.Error())
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
	t.Logf(err.Error())
}
