package config

import (
	"bauklotze/pkg/ioutils"
	"bauklotze/pkg/lockfile"
	"bauklotze/pkg/machine/env"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const connectionsFile = "bugbox-connections.json"

type ConnectionsFile struct {
	Connection ConnectionConfig `json:",omitempty"`
}

// connectionsConfigFile returns the path to the rw connections config file
func connectionsConfigFile() (string, error) {
	if path, found := os.LookupEnv("PODMAN_CONNECTIONS_CONF"); found {
		return path, nil
	}
	path, err := env.GetMachineConfDir()
	if err != nil {
		return "", err
	}
	// file is stored next to containers.conf
	p := filepath.Join(path, connectionsFile)
	return p, nil
}

func readConnectionConf(path string) (*ConnectionsFile, error) {
	conf := new(ConnectionsFile)
	f, err := os.Open(path)
	if err != nil {
		// return empty config if file does not exists
		if errors.Is(err, fs.ErrNotExist) {
			return conf, nil
		}

		return nil, err
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
		return err
	}

	opts := &ioutils.AtomicFileWriterOptions{ExplicitCommit: true}
	configFile, err := ioutils.NewAtomicFileWriterWithOpts(path, 0o644, opts)
	if err != nil {
		return err
	}
	defer configFile.Close()

	err = json.NewEncoder(configFile).Encode(conf)
	if err != nil {
		return err
	}

	// If no errors commit the changes to the config file
	return configFile.Commit()
}

func EditConnectionConfig(callback func(cfg *ConnectionsFile) error) error {
	path, err := connectionsConfigFile()
	if err != nil {
		return err
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
		return err
	}

	return writeConnectionConf(path, conf)
}
