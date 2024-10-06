package server

import (
	"github.com/containers/podman/v5/pkg/util"
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"
	"reflect"
	"syscall"
)

func NewAPIDecoder() *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)

	var Signal syscall.Signal
	d.RegisterConverter(Signal, convertSignal)
	return d
}

func convertSignal(query string) reflect.Value {
	signal, err := util.ParseSignal(query)
	if err != nil {
		logrus.Infof("convertSignal: Failed to parse %s: %s", query, err.Error())
	}
	return reflect.ValueOf(signal)
}
