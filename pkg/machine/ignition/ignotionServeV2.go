package ignition

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
)

var commandLists []string

func ignFileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "GET request received. Ignition files: %v\n", commandLists)
		logrus.Infof("GET request received. Ignition files: %v\n", commandLists)
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		if string(body) == "ServiceStop" {
			fmt.Fprintf(w, "POST request received. Stopping server.\n")
			logrus.Infof("POST request received. Stopping server.\n")
			stopChan <- struct{}{}
			return
		}
		fmt.Fprintf(w, "POST request received. Data: %s\n", body)
		logrus.Infof("POST request received. Data: %s\n", body)
	default:
		http.Error(w, "Unsupported request method", http.StatusMethodNotAllowed)
	}
}

var stopChan = make(chan struct{})

func IgnitionServeV2(commands []string) {
	commandLists = commands

	http.HandleFunc("/ignfile/", ignFileHandler)
	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stopChan
	_ = server.Close()
	log.Println("Server stopped")
}
