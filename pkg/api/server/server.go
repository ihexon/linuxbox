//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bauklotze/pkg/api/backend"
	"bauklotze/pkg/api/internal"
	"bauklotze/pkg/api/types"
	"bauklotze/pkg/machine/io"
	"bauklotze/pkg/machine/vmconfig"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	Server   http.Server
	Listener net.Listener
}

func RestService(ctx context.Context, mc *vmconfig.MachineConfig, endPoint string) error {
	// Set stdin to /dev/null
	_ = internal.RedirectStdin()
	// When deleting files, wrap the path in a `&io.FileWrapper` so that the file is safely deleted.
	// The Delete(true) operation will ensure that **only files in the workspace are deleted**
	UDF := &io.FileWrapper{Path: endPoint}
	if err := UDF.Delete(true); err != nil {
		return fmt.Errorf("failed to delete file %q: %w", UDF.GetPath(), err)
	}

	u := url.URL{
		Scheme: "unix",
		Path:   UDF.GetPath(),
	}

	listener, err := net.Listen(u.Scheme, u.Path)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", u.Path, err)
	}

	if !UDF.Exist() {
		return errors.New("UDF file create failed")
	}

	server := makeNewServer(mc, listener)
	defer func() {
		if err := server.Shutdown(); err != nil {
			logrus.Warnf("error when stopping API service: %s", err)
			_ = server.Close()
		}
	}()

	go func() {
		<-ctx.Done()
		logrus.Warnf("API service is shutting down duto error: %s", context.Cause(ctx))
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

func makeNewServer(mc *vmconfig.MachineConfig, listener net.Listener) *APIServer {
	router := mux.NewRouter().UseEncodedPath()

	server := APIServer{
		Server: http.Server{
			Handler: router, // Mux
		},
		Listener: listener,
	}

	server.Server.BaseContext = func(l net.Listener) context.Context {
		// Every request will have access to the machineConfig,this is a way to pass the machineConfig to the handlers
		ctx := context.WithValue(context.Background(), types.McKey, mc)
		return ctx
	}

	router.Use(PanicHandler())
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// We can track user errors...
			logrus.Warnf("RESTAPI Request failed: (%d:%s) for %s:'%s'", http.StatusNotFound, http.StatusText(http.StatusNotFound), r.Method, r.URL.String())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
	)

	server.setupRouter(router)

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

	return &server
}

func (s *APIServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx) //nolint:wrapcheck
}

func (s *APIServer) setupRouter(r *mux.Router) *mux.Router {
	r.Handle("/{name}/info", s.APIHandler(backend.GetInfos)).Methods(http.MethodGet)
	r.Handle("/{name}/exec", s.APIHandler(backend.DoExec)).Methods(http.MethodPost)
	return r
}

// Close immediately stops responding to clients and exits
func (s *APIServer) Close() error {
	return s.Server.Close() //nolint:wrapcheck
}
