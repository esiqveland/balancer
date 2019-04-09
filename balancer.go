package balancer

import (
	"net"

	"github.com/pkg/errors"
)

var ErrNoHosts = errors.New("No hosts available in list")

type Balancer interface {
	Next() (Host, error)
}

type Host struct {
	Address net.IP
	Port    int
}
