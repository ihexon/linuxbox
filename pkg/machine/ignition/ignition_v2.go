//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"

	ignition "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/sirupsen/logrus"
)

const (
	DefaultIgnitionUserName = "root"
	DefaultIgnitionVersion  = "3.4.0"
	GenerateScriptDir       = "/.ignfiles"
	PodmanMachine           = "/etc/containers/podman-machine"
	SystemEtcDir            = "/etc"
	SystemOpenRcDir         = "/etc/init.d/"
	SystemDefaultRunlevels  = "/etc/runlevels/default"
)

type DynamicIgnitionV2 struct {
	Name           string // vm user, default is root
	Key            string // sshkey
	TimeZone       string
	UID            int
	VMName         string
	VMType         define.VMType
	MachineConfigs *vmconfigs.MachineConfig
	WritePath      string
	Cfg            ignition.Config
	Rootful        bool
	NetRecover     bool
	Rosetta        bool
}

func (ign *DynamicIgnitionV2) Write() error {
	b, err := json.Marshal(ign.Cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal ignition config: %w", err)
	}
	if err := os.WriteFile(ign.WritePath, b, 0644); err != nil {
		return fmt.Errorf("failed to write ignition file: %w", err)
	}

	return nil
}

// Convenience function to convert int to ptr
//
// See: https://coreos.github.io/ignition/configuration-v3_4/
func IntToPtr(i int) *int {
	return &i
}

func GetNodeGrp(grpName string) ignition.NodeGroup {
	return ignition.NodeGroup{Name: &grpName}
}

func GetNodeUsr(usrName string) ignition.NodeUser {
	return ignition.NodeUser{Name: &usrName}
}

// Convenience function to convert bool to ptr
func BoolToPtr(b bool) *bool {
	return &b
}

func StrToPtr(s string) *string {
	return &s
}

func EncodeDataURLPtr(contents string) *string {
	return StrToPtr(fmt.Sprintf("data:,%s", url.PathEscape(contents)))
}

// getDirs: create directories
func (ign *DynamicIgnitionV2) getDirs(usrName string) []ignition.Directory {
	newDirs := []string{
		GenerateScriptDir,
	}

	dirs := make([]ignition.Directory, len(newDirs))

	for i, d := range newDirs {
		dirs[i] = ignition.Directory{
			Node: ignition.Node{
				Overwrite: BoolToPtr(true),
				Group:     GetNodeGrp(usrName),
				User:      GetNodeUsr(usrName),
				Path:      d,
			},
			DirectoryEmbedded1: ignition.DirectoryEmbedded1{Mode: IntToPtr(0755)},
		}
	}
	return dirs
}

// getUsers: Set sshkeys for root user
func (ign *DynamicIgnitionV2) getUsers() []ignition.PasswdUser {
	var (
		// See https://coreos.github.io/ignition/configuration-v3_4/
		users []ignition.PasswdUser
	)

	// set root SSH key
	root := ignition.PasswdUser{
		Name:              DefaultIgnitionUserName,
		SSHAuthorizedKeys: []ignition.SSHAuthorizedKey{ignition.SSHAuthorizedKey(ign.Key)},
	}
	// add them all in
	users = append(users, root)
	return users
}

func (ign *DynamicIgnitionV2) getFiles(usrName string, _uid int, vmtype define.VMType) []ignition.File {
	files := make([]ignition.File, 0)

	testConfigure := `# Test configures for root user`
	// Set test.conf up for root user, just a test
	files = append(files, ignition.File{
		Node: ignition.Node{
			Group: GetNodeGrp(usrName),
			Path:  filepath.Join(GenerateScriptDir, "test.conf"),
			User:  GetNodeUsr(usrName),
		},
		FileEmbedded1: ignition.FileEmbedded1{
			Append: nil,
			Contents: ignition.Resource{
				Source: EncodeDataURLPtr(testConfigure),
			},
			Mode: IntToPtr(0744),
		},
	})

	subUID := 100000
	subUIDs := 1000000
	etcSubUID := fmt.Sprintf(`%s:%d:%d`, usrName, subUID, subUIDs)

	// Set up /etc/subuid and /etc/subgid,
	// Podman need this see https://wiki.alpinelinux.org/wiki/Podman
	for _, sub := range []string{"/etc/subuid", "/etc/subgid"} {
		files = append(files, ignition.File{
			Node: ignition.Node{
				Group:     GetNodeGrp("root"),
				Path:      sub,
				User:      GetNodeUsr("root"),
				Overwrite: BoolToPtr(true),
			},
			FileEmbedded1: ignition.FileEmbedded1{
				Append: nil,
				Contents: ignition.Resource{
					Source: EncodeDataURLPtr(etcSubUID),
				},
				Mode: IntToPtr(0744),
			},
		})
	}

	// write libkrun to /etc/containers/podman-machine
	files = append(files, ignition.File{
		Node: ignition.Node{
			Group: GetNodeGrp(DefaultIgnitionUserName),
			Path:  PodmanMachine,
			User:  GetNodeUsr(DefaultIgnitionUserName),
		},
		FileEmbedded1: ignition.FileEmbedded1{
			Append: nil,
			Contents: ignition.Resource{
				Source: EncodeDataURLPtr(fmt.Sprintf("%s\n", vmtype.String())),
			},
			Mode: IntToPtr(0644),
		},
	})

	virtioMountRCFiles := ign.generateVirtioMountRC()
	files = append(files, virtioMountRCFiles...)

	return files
}

func (ign *DynamicIgnitionV2) getLinks(_usrName string) []ignition.Link {
	links := []ignition.Link{
		{
			Node: ignition.Node{
				Group:     GetNodeGrp(DefaultIgnitionUserName),
				Path:      "/usr/local/bin/docker",
				Overwrite: BoolToPtr(true),
				User:      GetNodeUsr(DefaultIgnitionUserName),
			},
			LinkEmbedded1: ignition.LinkEmbedded1{
				Hard:   BoolToPtr(false),
				Target: StrToPtr("/usr/bin/podman"),
			},
		},
	}

	for _, vol := range ign.MachineConfigs.Mounts {
		links = append(links, ignition.Link{
			Node: ignition.Node{
				Path:      filepath.Join(SystemDefaultRunlevels, rcPrefix+vol.Tag),
				Overwrite: BoolToPtr(true),
				User:      GetNodeUsr(DefaultIgnitionUserName),
				Group:     GetNodeGrp(DefaultIgnitionUserName),
			},
			LinkEmbedded1: ignition.LinkEmbedded1{
				Hard:   BoolToPtr(false),
				Target: StrToPtr(filepath.Join(SystemOpenRcDir, rcPrefix+vol.Tag)),
			},
		})
	}

	return links
}

// :(
func (ign *DynamicIgnitionV2) getAllMounts() []ignition.Filesystem {
	fs := make([]ignition.Filesystem, 0)

	return fs
}

type IgnitionBuilder struct {
	dynamicIgnition DynamicIgnitionV2
}

func NewIgnitionBuilder(dynamicIgnition DynamicIgnitionV2) IgnitionBuilder {
	return IgnitionBuilder{
		dynamicIgnition,
	}
}

func (ign *DynamicIgnitionV2) GenerateIgnitionConfig() error {
	if len(ign.Name) < 1 {
		ign.Name = DefaultIgnitionUserName
	}

	ignVersion := ignition.Ignition{
		Version: DefaultIgnitionVersion,
	}

	ignPasswd := ignition.Passwd{
		Users: ign.getUsers(),
	}

	ignStorage := ignition.Storage{
		Filesystems: ign.getAllMounts(),
		Directories: ign.getDirs(ign.Name),
		Files:       ign.getFiles(ign.Name, ign.UID, ign.VMType),
		Links:       ign.getLinks(ign.Name),
	}

	if len(ign.TimeZone) > 0 {
		var (
			err error
			tz  string
		)
		// local means the same as the host
		// look up where it is pointing to on the host
		if ign.TimeZone == "local" {
			tz, err = getLocalTimeZone()
			if err != nil {
				return err
			}
		} else {
			tz = ign.TimeZone
		}
		tzLink := ignition.Link{
			Node: ignition.Node{
				Group:     GetNodeGrp("root"),
				Path:      "/etc/localtime",
				Overwrite: BoolToPtr(true),
				User:      GetNodeUsr("root"),
			},
			LinkEmbedded1: ignition.LinkEmbedded1{
				Hard: BoolToPtr(false),
				// We always want this value in unix form (/path/to/something) because this is being
				// set in the machine OS (always Linux).  However, filepath.join on windows will use a "\\"
				// separator; therefore we use ToSlash to convert the path to unix style
				Target: StrToPtr(filepath.ToSlash(filepath.Join("/usr/share/zoneinfo", tz))),
			},
		}
		ignStorage.Links = append(ignStorage.Links, tzLink)
	}

	ign.Cfg = ignition.Config{
		Ignition: ignVersion,
		Passwd:   ignPasswd,
		Storage:  ignStorage,
	}

	return nil
}

// GenerateIgnitionConfig generates the ignition config
func (i *IgnitionBuilder) GenerateIgnitionConfig() error {
	return i.dynamicIgnition.GenerateIgnitionConfig()
}

func (i *IgnitionBuilder) WithFile(files ...ignition.File) {
	i.dynamicIgnition.Cfg.Storage.Files = append(i.dynamicIgnition.Cfg.Storage.Files, files...)
}

func (i *IgnitionBuilder) BuildWithIgnitionFile(ignPath string) error {
	inputIgnition, err := os.ReadFile(ignPath)
	if err != nil {
		return fmt.Errorf("failed to read ignition file: %w", err)
	}

	if err := os.WriteFile(i.dynamicIgnition.WritePath, inputIgnition, 0644); err != nil {
		return fmt.Errorf("failed to write ignition file: %w", err)
	}

	return nil
}

func (i *IgnitionBuilder) Build() error {
	logrus.Infof("writing ignition file to %q", i.dynamicIgnition.WritePath)
	return i.dynamicIgnition.Write()
}

var (
	virtioFsMountRc *bytes.Buffer
)

func (ign *DynamicIgnitionV2) generateVirtioMountRC() []ignition.File {
	virtioFsOpenRcCfg := make([]ignition.File, 0)

	for _, vol := range ign.MachineConfigs.Mounts {
		virtioFsMountRc = new(bytes.Buffer)

		sourceDev := vol.Tag
		targetPath := vol.Target
		fsType := vol.Type

		if fsType != virtiofs {
			continue
		}

		data := struct {
			FsType string
			Source string
			Target string
		}{
			FsType: fsType,
			Source: sourceDev,
			Target: targetPath,
		}
		t := template.Must(template.New("VirtioFsMountRcFile").Parse(MountOpenrcTemplate))
		_ = t.Execute(virtioFsMountRc, data)

		virtioFsOpenRcCfg = append(virtioFsOpenRcCfg, ignition.File{
			Node: ignition.Node{
				Group: GetNodeGrp(DefaultIgnitionUserName),
				User:  GetNodeUsr(DefaultIgnitionUserName),
				Path:  filepath.Join(SystemOpenRcDir, rcPrefix+vol.Tag),
			},

			FileEmbedded1: ignition.FileEmbedded1{
				Append: nil,
				Contents: ignition.Resource{
					Source: EncodeDataURLPtr(virtioFsMountRc.String()),
				},
				Mode: IntToPtr(0644),
			},
		})
	}

	return virtioFsOpenRcCfg
}
