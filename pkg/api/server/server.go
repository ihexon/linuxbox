package server

import (
	"bauklotze/pkg/api/server/internal"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"net"
	"net/http"
	"net/url"
	"os"

	"time"
)

func RestService(flags *pflag.FlagSet, apiurl string) error {
	var (
		listener net.Listener
		err      error
	)
	if apiurl == "" {
		if _, found := os.LookupEnv("BAZ_API_LISTEN_DIR"); !found {
			return errors.New("no service URI provided and socket activation protocol is not active")
		}
	}
	uri, err := url.Parse(apiurl)
	if err != nil {
		logrus.Errorf("%s is an invalid socket destination", apiurl)
		return fmt.Errorf("%s is an invalid socket destination", apiurl)
	}

	switch uri.Scheme {
	case "tcp":
		listener, err = net.Listen("tcp", uri.Host)
		if err != nil {
			logrus.Errorf("Failed to listen on %s: %s", uri.Host, err)
			return fmt.Errorf("Failed to listen on %s: %w", uri.Host, err)
		}
	default:
		return fmt.Errorf("API Service endpoint scheme %q is not supported. Try tcp://%s", uri.Scheme, apiurl)
	}

	for _, val := range []string{"LISTEN_FDS", "LISTEN_PID", "LISTEN_FDNAMES", "BAZ_API_LISTEN_DIR"} {
		if err := os.Unsetenv(val); err != nil {
			return fmt.Errorf("unsetting %s: %v", val, err)
		}
	}

	_ = internal.RedirectStdin()
	server, _ := makeNewServer(listener)

	defer func() {
		if err := server.Shutdown(); err != nil {
			logrus.Warnf("Error when stopping API service: %s", err)
		}
	}()

	err = server.Serve()
	if listener != nil {
		_ = listener.Close()
	}

	return err
}

func (s *APIServer) Serve() error {
	errChan := make(chan error, 1)
	go func() {
		err := s.Server.Serve(s.Listener)
		if err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start API service: %w", err)
			return
		}
		errChan <- nil
	}()
	return <-errChan
}

func makeNewServer(listener net.Listener) (*APIServer, error) {
	logrus.Infof("API service listening on %q. URI: %q", listener.Addr())
	router := mux.NewRouter().UseEncodedPath()

	tracker := newIdleTracker()
	server := APIServer{
		Server: http.Server{
			ConnState: tracker.ConnState,
			Handler:   router,
		},
		Listener:    listener,
		idleTracker: tracker,
	}

	server.BaseContext = func(l net.Listener) context.Context {
		return context.Background()
	}

	router.Use(PanicHandler(), ReferenceIDHandler())
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// We can track user errors...
			logrus.Infof("Failed Request: (%d:%s) for %s:'%s'", http.StatusNotFound, http.StatusText(http.StatusNotFound), r.Method, r.URL.String())
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
	)

	router.MethodNotAllowedHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// We can track user errors...
			logrus.Infof("Failed Request: (%d:%s) for %s:'%s'", http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), r.Method, r.URL.String())
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		},
	)

	setupRouter(router)

	return &server, nil
}

type APIServer struct {
	http.Server
	net.Listener
	context.Context
	context.CancelFunc
	idleTracker *idleTracker
}

func (s *APIServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

func setupRouter(router *mux.Router) *mux.Router {

	router.Handle("/version", http.HandlerFunc(versionHandler)).Methods(http.MethodGet)
	router.Handle("/getversion", http.HandlerFunc(versionHandler)).Methods(http.MethodGet)
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return router
}
