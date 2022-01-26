package balancer

import (
	"errors"
	"net"
	"strconv"
)

// ErrNoHosts if the lookup returned 0 hosts.
var ErrNoHosts = errors.New("No hosts available in list")

// Balancer returns the next Host to use.
type Balancer interface {
	Next() (Host, error)
}

type Host struct {
	Address net.IP
	Port    int
}

func (h Host) String() string {
	return h.Address.String() + ":" + strconv.Itoa(h.Port)
}
