package server

import (
	"github.com/gorilla/schema"
)

func NewAPIDecoder() *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)

	return d
}
