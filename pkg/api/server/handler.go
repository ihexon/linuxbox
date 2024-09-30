package server

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"runtime"
)

type APIContextKey int

const (
	DecoderKey APIContextKey = iota
	RuntimeKey
	IdleTrackerKey
	ConnKey
	CompatDecoderKey
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// panicHandler captures panics from endpoint handlers and logs stack trace
func PanicHandler() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// http.Server hides panics from handlers, we want to record them and fix the cause
			defer func() {
				err := recover()
				if err != nil {
					buf := make([]byte, 1<<20)
					n := runtime.Stack(buf, true)
					logrus.Warnf("Recovering from API service endpoint handler panic: %v, %s", err, buf[:n])
					// Try to inform client things went south... won't work if handler already started writing response body
					InternalServerError(w, fmt.Errorf("%v", err))
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}

func InternalServerError(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, err)
}

func Error(w http.ResponseWriter, code int, err error) {
	// Log detailed message of what happened to machine running podman service

	em := ErrorModel{
		Because:      Cause(err).Error(),
		Message:      err.Error(),
		ResponseCode: code,
	}
	WriteJSON(w, code, em)
}

func ReferenceIDHandler() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		// Only log Apache access_log-like entries at Info level or below
		out := io.Discard
		if logrus.IsLevelEnabled(logrus.InfoLevel) {
			out = logrus.StandardLogger().Out
		}

		return handlers.CombinedLoggingHandler(out,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rid := r.Header.Get("X-Reference-Id")
				if rid == "" {
					if c := r.Context().Value(ConnKey); c == nil {
						rid = uuid.New().String()
					} else {
						rid = fmt.Sprintf("%p", c)
					}
				}

				r.Header.Set("X-Reference-Id", rid)
				w.Header().Set("X-Reference-Id", rid)
				h.ServeHTTP(w, r)
			}))
	}
}

func WriteJSON(w http.ResponseWriter, code int, value interface{}) {
	// FIXME: we don't need to write the header in all/some circumstances.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	coder := json.NewEncoder(w)
	coder.SetEscapeHTML(true)
	if err := coder.Encode(value); err != nil {
		logrus.Errorf("Unable to write json: %q", err)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
