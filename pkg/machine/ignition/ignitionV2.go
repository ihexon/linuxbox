package ignition

import (
	"bauklotze/pkg/machine/define"
	"encoding/json"
	"fmt"
	ignition "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"path/filepath"
)

type DynamicIgnitionV2 struct {
	Name       string // vm user, default is root
	Key        string // sshkey
	TimeZone   string
	UID        int
	VMName     string
	VMType     define.VMType
	WritePath  string
	Cfg        ignition.Config
	Rootful    bool
	NetRecover bool
	Rosetta    bool
}

func (ign *DynamicIgnitionV2) Write() error {
	b, err := json.Marshal(ign.Cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ign.WritePath, b, 0644)
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

const (
	DefaultIgnitionUserName = "root"
	DefaultIgnitionVersion  = "3.4.0"
)

func EncodeDataURLPtr(contents string) *string {
	return StrToPtr(fmt.Sprintf("data:,%s", url.PathEscape(contents)))
}

func getDirs(usrName string) []ignition.Directory {
	// Ignition has a bug/feature? where if you make a series of dirs
	// in one swoop, then the leading dirs are creates as root.
	newDirs := []string{
		"/" + usrName + "/.config",
	}
	var (
		dirs = make([]ignition.Directory, len(newDirs))
	)
	for i, d := range newDirs {
		newDir := ignition.Directory{
			Node: ignition.Node{
				Group: GetNodeGrp(usrName),
				Path:  d,
				User:  GetNodeUsr(usrName),
			},
			DirectoryEmbedded1: ignition.DirectoryEmbedded1{Mode: IntToPtr(0755)},
		}
		dirs[i] = newDir
	}

	return dirs
}

func (ign *DynamicIgnitionV2) getUsers() []ignition.PasswdUser {
	var (
		// See https://coreos.github.io/ignition/configuration-v3_4/
		users []ignition.PasswdUser
	)

	// set root SSH key
	root := ignition.PasswdUser{
		Name:              ign.Name,
		SSHAuthorizedKeys: []ignition.SSHAuthorizedKey{ignition.SSHAuthorizedKey(ign.Key)},
	}
	// add them all in
	users = append(users, root)
	return users
}

func getFiles(usrName string, uid int, rootful bool, vmtype define.VMType, _ bool) []ignition.File {
	files := make([]ignition.File, 0)

	containers := `
#netns="bridge"
#pids_limit=0
`
	subUID := 100000
	subUIDs := 1000000
	if uid >= subUID && uid < (subUID+subUIDs) {
		subUID = uid + 1
	}
	etcSubUID := fmt.Sprintf(`%s:%d:%d`, usrName, subUID, subUIDs)

	// Set test.conf up for root user, just a test
	files = append(files, ignition.File{
		Node: ignition.Node{
			Group: GetNodeGrp(usrName),
			Path:  "/" + usrName + "/.config/test.conf",
			User:  GetNodeUsr(usrName),
		},
		FileEmbedded1: ignition.FileEmbedded1{
			Append: nil,
			Contents: ignition.Resource{
				Source: EncodeDataURLPtr(containers),
			},
			Mode: IntToPtr(0744),
		},
	})

	// Set up /etc/subuid and /etc/subgid
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

	// Set machine marker file to indicate podman what vmtype we are
	// operating under
	files = append(files, ignition.File{
		Node: ignition.Node{
			Group: GetNodeGrp(DefaultIgnitionUserName),
			Path:  "/etc/containers/podman-machine",
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

	return files
}

func getLinks(usrName string) []ignition.Link {
	return []ignition.Link{
		{
			Node: ignition.Node{
				Group:     GetNodeGrp("root"),
				Path:      "/usr/local/bin/docker",
				Overwrite: BoolToPtr(true),
				User:      GetNodeUsr("root"),
			},
			LinkEmbedded1: ignition.LinkEmbedded1{
				Hard:   BoolToPtr(false),
				Target: StrToPtr("/usr/bin/podman"),
			},
		},
	}
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
		Directories: getDirs(ign.Name),
		Files:       getFiles(ign.Name, ign.UID, ign.Rootful, ign.VMType, ign.NetRecover),
		Links:       getLinks(ign.Name),
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
				Overwrite: BoolToPtr(false),
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
		return err
	}

	return os.WriteFile(i.dynamicIgnition.WritePath, inputIgnition, 0644)
}

func (i *IgnitionBuilder) Build() error {
	logrus.Infof("writing ignition file to %q", i.dynamicIgnition.WritePath)
	return i.dynamicIgnition.Write()
}
