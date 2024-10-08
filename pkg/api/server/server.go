package server

import (
	"bauklotze/pkg/api/backend"
	"bauklotze/pkg/api/internal"
	"bauklotze/pkg/api/types"
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

type APIServer struct {
	http.Server
	net.Listener
	context.Context
	context.CancelFunc
	CorsHeaders string // Inject Cross-Origin Resource Sharing (CORS) headers
	idleTracker *idleTracker
}

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
		return fmt.Errorf("%s is an invalid socket destination", apiurl)
	}

	switch uri.Scheme {
	case "tcp":
		listener, err = net.Listen(uri.Scheme, uri.Host)
		if err != nil {
			return fmt.Errorf("Failed to listen on %s: %w", uri.Host, err)
		}
	default:
		return fmt.Errorf("API Service endpoint scheme %q is not supported. Try tcp://%s", uri.Scheme, apiurl)
	}

	// Disable leaking the LISTEN_* into containers
	for _, val := range []string{"LISTEN_FDS", "LISTEN_PID", "LISTEN_FDNAMES", "BAZ_API_LISTEN_DIR"} {
		if err := os.Unsetenv(val); err != nil {
			return fmt.Errorf("unsetting %s: %v", val, err)
		}
	}

	// Set stdin to /dev/null
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

// Serve is the wrapper of http.Server.Serve, will block the code path until the server stopping or getting error.
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
	server.setupRouter(router)
	return &server, nil
}

func (s *APIServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

func (s *APIServer) setupRouter(r *mux.Router) *mux.Router {
	r.Handle(("/apiversion"), s.APIHandler(backend.VersionHandler)).Methods(http.MethodGet)
	r.Handle(("/{name}/diskuage"), s.APIHandler(backend.GetDiskUsage)).Methods(http.MethodGet)
	r.Handle(("/{name}/info"), s.APIHandler(backend.GetInfos)).Methods(http.MethodGet)
	return r
}
