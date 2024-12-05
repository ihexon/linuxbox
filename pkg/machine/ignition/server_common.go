//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package ignition

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// ServeIgnitionOverSocketCommon Is A block function, design to be running in go routine
func ServeIgnitionOverSocketCommon(url *url.URL, file fs.File) error {
	listener, err := net.Listen(url.Scheme, url.Path)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", url.Path, err)
	}

	ignFile, err := io.ReadAll(file) // Ignition json file
	if err != nil {
		return fmt.Errorf("failed to read ignition file: %w", err)
	}

	s, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stat: %w", err)
	}
	cfgAbsPath, err := filepath.Abs(s.Name())
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %s, %w", s.Name(), err)
	}

	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("New request : %s", r.RequestURI)
		logrus.Infof("Serving ignition file: %s", cfgAbsPath)
		_, err := w.Write(ignFile) // Ignition json file
		if err != nil {
			logrus.Errorf("Failed to serve ignition file: %v", err)
		}
	})

	mux.HandleFunc("/fetch/stop", func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("Serving ignition file: %s", cfgAbsPath)
		_, err := w.Write(ignFile) // Ignition json file
		if err != nil {
			logrus.Errorf("Failed to serve ignition file: %v", err)
		}
		errChan <- fmt.Errorf("fetch %s and stop", cfgAbsPath)
	})

	logrus.Infof("Ignition listening on: %s:%s/%s", url.Scheme, url.Host, url.Path)

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("Failed to serve ignition file: %v", err)
			errChan <- err
		}
	}()

	err = <-errChan // BLOCKED
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Ignition server Shutdown error: %v", context.Cause(shutdownCtx))
	} else {
		logrus.Infof("Ignition server Shutdown successful")
	}

	return err
}
