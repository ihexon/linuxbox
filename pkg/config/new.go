package config

import "sync"

// Options to use when loading a Config via New().
type Options struct {
	SetDefault bool
}

var (
	cachedConfigMutex sync.Mutex
)

// New returns a Config as described in the containers.conf(5) man page.
func New(options *Options) (*Config, error) {
	if options == nil {
		options = &Options{}
	} else if options.SetDefault {
		cachedConfigMutex.Lock()
		defer cachedConfigMutex.Unlock()
	}
	return newLocked(options)
}

func newLocked(options *Options) (*Config, error) {
	// Start with the built-in defaults
	config, err := defaultConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}
