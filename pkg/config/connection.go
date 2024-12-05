//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"bauklotze/pkg/machine/env"

	"github.com/containers/storage/pkg/ioutils"
	"github.com/containers/storage/pkg/lockfile"
)

const connectionsFile = "connections.json"

type ConnectionConfig struct {
	Default     string                 `json:",omitempty"`
	Connections map[string]Destination `json:",omitempty"`
}

type ConnectionsFile struct {
	Connection ConnectionConfig `json:",omitempty"`
}

// connectionsConfigFile returns the path to the rw connections config file
func connectionsConfigFile() (string, error) {
	path, err := env.ConfDirPrefix()
	if err != nil {
		return "", fmt.Errorf("failed to get conf dir prefix: %w", err)
	}
	return filepath.Join(path, "connectionCfg", connectionsFile), nil // ${BauklotzeHomePath}/config/connectionCfg/bugbox-connections.json
}

func readConnectionConf(path string) (*ConnectionsFile, error) {
	conf := new(ConnectionsFile)
	f, err := os.Open(path)
	if err != nil {
		// return empty config if file does not exists
		if errors.Is(err, fs.ErrNotExist) {
			return conf, nil
		}

		return nil, fmt.Errorf("failed to open connections config: %w", err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(conf)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}
	return conf, nil
}

func writeConnectionConf(path string, conf *ConnectionsFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", filepath.Dir(path), err)
	}

	opts := &ioutils.AtomicFileWriterOptions{ExplicitCommit: true}
	configFile, err := ioutils.NewAtomicFileWriterWithOpts(path, 0o644, opts)
	if err != nil {
		return fmt.Errorf("create atomic file writer: %s, %w", path, err)
	}
	defer configFile.Close()

	err = json.NewEncoder(configFile).Encode(conf)
	if err != nil {
		return fmt.Errorf("failed to encode connections config: %s, %w", configFile, err)
	}

	// If no errors commit the changes to the config file
	return configFile.Commit() //nolint:wrapcheck
}

// EditConnectionConfig must be used to edit the connections config.
// The function will read and write the file automatically and the
// callback function just needs to modify the cfg as needed.
func EditConnectionConfig(callback func(cfg *ConnectionsFile) error) error {
	path, err := connectionsConfigFile()
	if err != nil {
		return fmt.Errorf("get connections config file path: %w", err)
	}

	lockPath := path + ".lock"
	lock, err := lockfile.GetLockFile(lockPath)
	if err != nil {
		return fmt.Errorf("obtain lock file: %w", err)
	}
	lock.Lock()
	defer lock.Unlock()

	conf, err := readConnectionConf(path)
	if err != nil {
		return fmt.Errorf("read connections file: %w", err)
	}

	if err := callback(conf); err != nil {
		return fmt.Errorf("failed to call edit connections config: %w", err)
	}

	return writeConnectionConf(path, conf)
}
