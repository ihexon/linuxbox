package handlers

import (
	"bauklotze/pkg/api/server"
	"bauklotze/pkg/api/server/utils"
	"net/http"
	"runtime"
)

const (
	unstable string = "unstable"
)

type Version struct {
	APIVersion string
	Version    string
	GoVersion  string
	OsArch     string
	Os         string
}

func getVersion() (Version, error) {
	return Version{
		APIVersion: unstable,
		Version:    unstable,
		GoVersion:  runtime.Version(),
		OsArch:     runtime.GOOS + "/" + runtime.GOARCH,
		Os:         runtime.GOOS,
	}, nil
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	running, err := getVersion()
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteResponse(w, http.StatusOK, running)
}
