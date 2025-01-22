//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package callback

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

type CleanupCallback struct {
	funcs []func() error
	mu    sync.Mutex
}

func (c *CleanupCallback) CleanIfErr(err *error) {
	// Do not remove created files if the init is successful
	if *err == nil {
		return
	}
	c.clean()
}

func (c *CleanupCallback) CleanOnSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sigType, ok := <-ch
	if !ok {
		return
	}
	logrus.Infof("Callback: clean up when catch sgnal %s", sigType.String())
	c.clean()
}

func (c *CleanupCallback) clean() {
	c.mu.Lock()
	// Claim exclusive usage by copy and resetting to nil
	funcs := c.funcs
	c.funcs = nil
	c.mu.Unlock()

	// Already claimed or none set
	if funcs == nil {
		return
	}

	// Cleanup functions can now exclusively be run
	for _, fn := range funcs {
		if err := fn(); err != nil {
			logrus.Errorf("callback fn() failed: %v", err.Error())
		}
	}
}

func CleanUp() CleanupCallback {
	return CleanupCallback{
		funcs: []func() error{},
	}
}

func (c *CleanupCallback) Add(anotherfunc func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, anotherfunc)
	c.mu.Unlock()
}
