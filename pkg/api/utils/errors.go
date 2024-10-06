package utils

import (
	"github.com/containers/podman/v5/pkg/errorhandling"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"net/http"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
