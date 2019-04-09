package srvbalancer

import (
	"net"

	"github.com/benschw/srv-lb/lb"
	"github.com/esiqveland/balancer"
	"github.com/pkg/errors"
)

// NewSRVBalancer creates a new balancer based on lookup of DNS SRV records.
// example usage: name: "_http", proto: "_tcp", host: "backend.namespace.kube.dc.org"
func NewSRVBalancer(name, proto, host string) (balancer.Balancer, error) {
	srvName := name + "." + proto + "." + host

	cfg, err := lb.DefaultConfig()
	if err != nil {
		return nil, err
	}

	l := lb.New(cfg, srvName)

	_, err = l.Next()
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up host=%v", srvName)
	}

	bal := &srvBalancer{
		balancer: l,
	}

	return bal, nil
}

type srvBalancer struct {
	balancer lb.LoadBalancer
}

func (lb *srvBalancer) Next() (balancer.Host, error) {
	addr, err := lb.balancer.Next()
	if err != nil {
		return balancer.Host{}, err
	}

	ip := net.ParseIP(addr.Address)
	if ip == nil {
		return balancer.Host{}, errors.Errorf("unable to parse ip=%v", addr.Address)
	}

	host := balancer.Host{
		Address: ip,
		Port:    int(addr.Port),
	}

	return host, nil
}
