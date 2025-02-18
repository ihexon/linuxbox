//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package server

import (
	"bufio"
	"fmt"
	"net/http"
	"runtime"

	"github.com/containers/podman/v5/pkg/errorhandling"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type APIContextKey int

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// PanicHandler captures panics from endpoint handlers and logs stack trace
func PanicHandler() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// http.Server hides panics from handlers, we want to record them and fix the cause
			defer func() {
				if err := recover(); err != nil {
					buf := make([]byte, 1<<20) //nolint:mnd
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

func WriteJSON(w http.ResponseWriter, code int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	coder := json.NewEncoder(w)
	coder.SetEscapeHTML(true)
	if err := coder.Encode(value); err != nil {
		logrus.Errorf("Unable to write json: %q", err)
	}
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
		s.apiWrapper(h, w, r)
	}
}

func (s *APIServer) apiWrapper(h http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logrus.Errorf("Failed Request: unable to parse form: " + err.Error())
	}
	h(w, r)
}

type BufferedResponseWriter struct {
	b *bufio.Writer
	w http.ResponseWriter
}

func (w *BufferedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *BufferedResponseWriter) Write(b []byte) (int, error) {
	return w.b.Write(b) //nolint:wrapcheck
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
