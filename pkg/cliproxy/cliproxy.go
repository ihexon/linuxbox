package cliproxy

import (
	"bauklotze/pkg/machine/env"
	"embed"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed resources/*
var embeddedFiles embed.FS

func RunCliProxy() (*exec.Cmd, error) {
	logrus.Infof("Running CliProxy....\n")
	var err error
	workspaceDir := os.Getenv(env.BAUKLOTZE_HOME)
	if workspaceDir == "" {
		return nil, fmt.Errorf("RunCliProxy(): %s is not set", env.BAUKLOTZE_HOME)
	}

	p := filepath.Join(workspaceDir) // p = ${BauklotzeHomePath}/

	logrus.Infof("Extracting cliproxy files into : %s\n", p)
	err = extractEmbeddedFiles(p)
	if err != nil {
		logrus.Errorf("Error extracting cliproxy files: %v\n", err)
	}

	DYLDLibraryPath := fmt.Sprintf("%s:%s", filepath.Join(p, "resources", "socat", "lib"), os.Getenv("DYLD_LIBRARY_PATH"))
	_ = os.Setenv(env.DYLD_LIBRARY_PATH, DYLDLibraryPath)

	const port = 5123
	r_side := fmt.Sprintf("TCP-LISTEN:%d,reuseaddr,fork", port)
	wrapper := filepath.Join(p, "resources", "socat", "bin", "wrapper") // wrapper = ${BauklotzeHomePath}/resources/socat/bin/wrapper
	l_side := fmt.Sprintf("EXEC:\"/bin/bash +x %s\",pty,echo=0,stderr", wrapper)
	timeout := "51840000" // just a big number that socat will wait for command execute finished

	// socat  -v -t 51840000 TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"bash +x wrapper",pty,echo=0,stderr
	socatBin := filepath.Join(p, "resources", "socat", "bin", "socat") // socatBin = ${BauklotzeHomePath}/resources/socat/bin/socat
	socatCmd := exec.Command(socatBin, "-v", "-t", timeout, r_side, l_side)
	socatCmd.Stdout = os.Stdout
	socatCmd.Stderr = os.Stderr

	logrus.Infof("Running socat: %v with DYLD_LIBRARY_PATH=%s \n", socatCmd, os.Getenv(env.DYLD_LIBRARY_PATH))
	err = socatCmd.Start()

	if err != nil {
		logrus.Errorf("Error running socat: %v\n", err)
		return nil, err
	}

	return socatCmd, err
}

func extractEmbeddedFiles(destination string) error {
	err := fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(destination, path)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0755)
	})

	return err
}
