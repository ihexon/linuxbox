package server

import (
	"bauklotze/pkg/api/types"
	"bufio"
	"fmt"
	"github.com/containers/podman/v5/pkg/errorhandling"
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

// A custom middleware
func ReferenceIDHandler() mux.MiddlewareFunc /* Note type MiddlewareFunc func(http.Handler) http.Handler */ {
	return func(h http.Handler) http.Handler { // 返回一个 http.Handler，实际上是返回 handlers.CombinedLoggingHandler
		// Only log Apache access_log-like entries at Info level or below
		out := io.Discard
		if logrus.IsLevelEnabled(logrus.InfoLevel) {
			out = logrus.StandardLogger().Out
		}

		// CombinedLoggingHandler 返回 loggingHandler struct 类型，而 loggingHandler struct 有自己的的
		// ServeHTTP(w http.ResponseWriter, req *http.Request) 方法。所以 CombinedLoggingHandler 返回的是一个实现了 http.Handler 接口的对象
		// type loggingHandler struct {
		//	writer    io.Writer --> out
		//	handler   http.Handler --> http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {...}
		//	formatter LogFormatter
		//}
		return handlers.CombinedLoggingHandler(out,
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					rid := r.Header.Get("X-Reference-Id")
					if rid == "" {
						if c := r.Context().Value(types.ConnKey); c == nil {
							rid = uuid.New().String()
						} else {
							rid = fmt.Sprintf("%p", c)
						}
					}

					r.Header.Set("X-Reference-Id", rid)
					w.Header().Set("X-Reference-Id", rid)
					h.ServeHTTP(w, r)
				},
			),
		)
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

func Error(w http.ResponseWriter, code int, err error) {
	// Log detailed message of what happened to machine running podman service
	logrus.Infof("Request Failed(%s): %s", http.StatusText(code), err.Error())
	em := errorhandling.ErrorModel{
		Because:      errorhandling.Cause(err).Error(),
		Message:      err.Error(),
		ResponseCode: code,
	}
	WriteJSON(w, code, em)
}

// APIHandler is a wrapper to enhance HandlerFunc's and remove redundant code
func (s *APIServer) APIHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Wrapper to hide some boilerplate
		s.apiWrapper(h, w, r, false)
	}
}

func (s *APIServer) apiWrapper(h http.HandlerFunc, w http.ResponseWriter, r *http.Request, buffer bool) {
	if err := r.ParseForm(); err != nil {
		logrus.WithFields(logrus.Fields{
			"X-Reference-Id": r.Header.Get("X-Reference-Id"),
		}).Info("Failed Request: unable to parse form: " + err.Error())
	}

	if s.CorsHeaders == "" {
		w.Header().Set("Access-Control-Allow-Origin", types.DefaultCORSAllowedHost)
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth, Connection, Upgrade, X-Registry-Config")
		w.Header().Set("Access-Control-Allow-Methods", "HEAD, GET, POST")
	}

	if buffer {
		bw := newBufferedResponseWriter(w)
		defer bw.b.Flush()
		w = bw
	}

	h(w, r)
}

type BufferedResponseWriter struct {
	b *bufio.Writer
	w http.ResponseWriter
}

func newBufferedResponseWriter(rw http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		bufio.NewWriterSize(rw, 8192),
		rw,
	}
}
func (w *BufferedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *BufferedResponseWriter) Write(b []byte) (int, error) {
	return w.b.Write(b)
}

func (w *BufferedResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

func (w *BufferedResponseWriter) Flush() {
	_ = w.b.Flush()
	if wrapped, ok := w.w.(http.Flusher); ok {
		wrapped.Flush()
	}
}
