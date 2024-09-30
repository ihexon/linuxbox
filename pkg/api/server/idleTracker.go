package server

import (
	"net"
	"net/http"
	"sync"
)

type idleTracker struct {
	mux               sync.Mutex
	activeConnections int
	totalConnections  int
}

func newIdleTracker() *idleTracker {
	return &idleTracker{
		totalConnections: 0,
	}
}

// Close is used to update Tracker that a StateHijacked connection has been closed by handler (StateClosed)
func (t *idleTracker) Close() {
	t.ConnState(nil, http.StateClosed)
}

func (t *idleTracker) TotalConnections() int {
	return t.totalConnections
}

func (t *idleTracker) ConnState(conn net.Conn, state http.ConnState) {
	t.mux.Lock()
	defer t.mux.Unlock()
	switch state {
	case http.StateNew:
		t.totalConnections++
	case http.StateActive:
		t.activeConnections++
	case http.StateClosed, http.StateHijacked:
		t.activeConnections--
	}
}
