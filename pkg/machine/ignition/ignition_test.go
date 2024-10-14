package ignition

import (
	"bauklotze/pkg/machine/define"
	"encoding/json"
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
		Name:      DefaultIgnitionUserName,
		Key:       "keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey",
		TimeZone:  "local", // Auto detect timezone from locales
		VMType:    define.LibKrun,
		VMName:    define.DefaultMachineName,
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
	err := ServeIgnitionOverSockV2(nil, nil)
	if err != nil {
		t.Fatalf("Failed to serve ignition over sock: %v", err)
	}
}
