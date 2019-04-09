package balancer

import (
	"errors"
	"net"
)

var ErrNoHosts = errors.New("No hosts available in list")

type Balancer interface {
	Next() (Host, error)
}

type Host struct {
	Address net.IP
	Port    int
}
