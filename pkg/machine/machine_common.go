package machine

import "sync"

var GlobalPIDs = &AllPIDs{}

type AllPIDs struct {
	KrunkitPID *int
	GvproxyPID *int
	mu         sync.Mutex
}

func (p *AllPIDs) SetKrunkitPID(pid int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.KrunkitPID = &pid
}

func (p *AllPIDs) GetKrunkitPID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return *p.KrunkitPID
}

func (p *AllPIDs) SetGvproxyPID(pid int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.GvproxyPID = &pid
}

func (p *AllPIDs) GetGvproxyPID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return *p.GvproxyPID
}

// TODO
func WaitAPIAndPrintInfo(forwardState APIForwardingState, name, forwardSock string, noInfo, rootful bool) {
}
