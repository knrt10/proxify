package proxy

import (
	"net"
	"strings"
)

type Target interface {
	// Address returns the address with which to access the target
	Address() string

	// Hostname returns the name of the host without address suffix
	Hostname() string

	// IsAlive returns true if the target is alive and able to serve requests
	IsAlive() bool
}

type targetServer struct {
	target string
}

func (t *targetServer) Address() string { return t.target }

func (t *targetServer) IsAlive() bool {
	hostname := t.Hostname()
	_, err := net.LookupHost(hostname)
	if err != nil {
		return false
	}

	return true
}

func (t *targetServer) Hostname() string {
	// Slice the hostname from the target string
	return strings.Split(t.target, ":")[0]
}

// NewTargetServer create a new targetserver instance
func NewTargetServer(target string) *targetServer {
	return &targetServer{target}
}
