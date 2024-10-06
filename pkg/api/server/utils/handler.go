package utils

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func WriteResponse(w http.ResponseWriter, code int, value interface{}) {
	// RFC2616 explicitly states that the following status codes "MUST NOT
	// include a message-body":
	switch code {
	case http.StatusNoContent, http.StatusNotModified: // 204, 304
		w.WriteHeader(code)
		return
	}

	switch v := value.(type) {
	case string:
		w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
		w.WriteHeader(code)

		if _, err := fmt.Fprintln(w, v); err != nil {
			logrus.Errorf("Unable to send string response: %q", err)
		}
	case *os.File:
		w.Header().Set("Content-Type", "application/octet; charset=us-ascii")
		w.WriteHeader(code)

		if _, err := io.Copy(w, v); err != nil {
			logrus.Errorf("Unable to copy to response: %q", err)
		}
	case io.Reader:
		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(code)

		if _, err := io.Copy(w, v); err != nil {
			logrus.Errorf("Unable to copy to response: %q", err)
		}
	default:
		WriteJSON(w, code, value)
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
