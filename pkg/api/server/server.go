//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bauklotze/pkg/api/backend"
	"bauklotze/pkg/api/internal"
	"bauklotze/pkg/api/types"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	http.Server
	net.Listener
	context.Context
	context.CancelFunc
	idleTracker *idleTracker
}

func RestService(ctx context.Context, apiurl *url.URL) error {
	var (
		listener net.Listener
		err      error
		path     string
	)

	switch apiurl.Scheme {
	case "unix":
		path, err = filepath.Abs(apiurl.Path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %s, %w", path, err)
		}
		if err = os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove all: %s, %w", path, err)
		}
		listener, err = net.Listen(apiurl.Scheme, path)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", path, err)
		}
	default:
		return fmt.Errorf("API Service endpoint scheme %q is not supported", apiurl.Scheme)
	}

	// Disable leaking the LISTEN_* into containers
	for _, val := range []string{"LISTEN_FDS", "LISTEN_PID", "LISTEN_FDNAMES", "BAZ_API_LISTEN_DIR"} {
		if err := os.Unsetenv(val); err != nil {
			return fmt.Errorf("unsetting %s: %w", val, err)
		}
	}

	// Set stdin to /dev/null
	_ = internal.RedirectStdin()
	server := makeNewServer(listener)

	defer func() {
		if err := server.Shutdown(); err != nil {
			logrus.Warnf("error when stopping API service: %s", err)
			_ = server.Close()
		}
	}()

	go func() {
		<-ctx.Done()
		logrus.Infof("API service is shutting down")
		if err := server.Shutdown(); err != nil {
			logrus.Warnf("error when stopping API service: %s", err)
			_ = server.Close()
		}
	}()

	return server.Serve()
}

// Serve is the wrapper of http.Server.Serve, will block the code path until the server stopping or getting error.
func (s *APIServer) Serve() error {
	errChan := make(chan error, 1)
	go func() {
		err := s.Server.Serve(s.Listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to start API service: %w", err)
			return
		}
		errChan <- nil
	}()
	return <-errChan
}

func makeNewServer(listener net.Listener) *APIServer {
	logrus.Infof("API service listening on %q.", listener.Addr())
	router := mux.NewRouter().UseEncodedPath()

	// setup a tracker to tracking every connections
	tracker := newIdleTracker()

	server := APIServer{
		Server: http.Server{
			ConnState: tracker.ConnState, // connection tracker
			Handler:   router,            // Mux
		},
		Listener:    listener,
		idleTracker: tracker,
	}

	server.Server.BaseContext = func(l net.Listener) context.Context {
		ctx := context.WithValue(context.Background(), types.DecoderKey, NewAPIDecoder()) // Decoder used in handlers as `decoder := r.Context().Value(api.DecoderKey).(*schema.Decoder)`
		return ctx
	}

	router.Use(PanicHandler(), ReferenceIDHandler())
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// We can track user errors...
			logrus.Warnf("RESTAPI Request failed: (%d:%s) for %s:'%s'", http.StatusNotFound, http.StatusText(http.StatusNotFound), r.Method, r.URL.String())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
	)

	server.setupRouter(router)

	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		_ = router.Walk(func(route *mux.Route, r *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				path = "<N/A>"
			}
			methods, err := route.GetMethods()
			if err != nil {
				methods = []string{"<N/A>"}
			}
			logrus.Infof("Methods: %6s Path: %s", strings.Join(methods, ", "), path)
			return nil
		})
	}

	return &server
}

func (s *APIServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx) //nolint:wrapcheck
}

func (s *APIServer) setupRouter(r *mux.Router) *mux.Router {
	r.Handle(("/apiversion"), s.APIHandler(backend.VersionHandler)).Methods(http.MethodGet)
	r.Handle(("/{name}/info"), s.APIHandler(backend.GetInfos)).Methods(http.MethodGet)
	r.Handle(("/{name}/vmstat"), s.APIHandler(backend.GetVMStat)).Methods(http.MethodGet)
	r.Handle(("/{name}/synctime"), s.APIHandler(backend.TimeSync)).Methods(http.MethodGet)
	r.Handle(("/{name}/exec"), s.APIHandler(backend.DoExec)).Methods(http.MethodPost)

	return r
}

// Close immediately stops responding to clients and exits
func (s *APIServer) Close() error {
	return s.Server.Close() //nolint:wrapcheck
}
