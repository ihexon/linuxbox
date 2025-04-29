//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package server

import (
	"bufio"
	"fmt"
	"net/http"
	"runtime"

	"bauklotze/pkg/api/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIContextKey int

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
					utils.Error(w, http.StatusInternalServerError, fmt.Errorf("%v", err))
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
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
		logrus.Errorf("Failed Request: unable to parse form: %v", err)
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
