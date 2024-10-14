package ignition

import (
	"bauklotze/pkg/machine/define"
	"bauklotze/pkg/machine/vmconfigs"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

// ServeIgnitionOverSock allows podman to open a small httpd instance on the vsock between the host
// and guest to inject the ignitionfile into fcos
func ServeIgnitionOverSockV2(listenedFile *define.VMFile, mc *vmconfigs.MachineConfig) error {
	ignitionFile, err := mc.IgnitionFile()
	var listener net.Listener

	logrus.Infof("reading ignition file: %s", ignitionFile.GetPath())
	ignFile, err := ignitionFile.Read()
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("serving ignition file....")
		_, err := w.Write(ignFile)
		if err != nil {
			logrus.Errorf("failed to serve ignition file: %v", err)
		}
	})

	//mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
	//	logrus.Infof("stop the ignition server....")
	//	if err := listener.Close(); err != nil {
	//		logrus.Errorf("failed to stop the server: %v", err)
	//	}
	//})

	//listenedFile := define.VMFile{Path: "/tmp/ignition1.sock"}
	_ = listenedFile.Delete()
	listener, err = net.Listen("unix", listenedFile.GetPath())
	if err != nil {
		return err
	}

	logrus.Infof("ignition socket device: %s", listenedFile.GetPath())
	defer func() {
		if err := listener.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	return http.Serve(listener, mux)
}
