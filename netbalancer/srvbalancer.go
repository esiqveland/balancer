package netbalancer

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/esiqveland/balancer"
)

type dnsSrvBalancer struct {
	serviceName string
	proto       string
	host        string
	hosts       []balancer.Host
	counter     uint64
	interval    time.Duration
	quit        chan int
	lock        *sync.Mutex
	Timeout     time.Duration
}

// NewSRV returns a Balancer that uses dns lookups from net.LookupSRV to reload a set of hosts every updateInterval.
// We can not use TTL from dns because TTL is not exposed by the Go stdlib.
func NewSRV(servicename, proto, host string, updateInterval time.Duration, dnsTimeout time.Duration) (balancer.Balancer, error) {
	initialHosts, err := lookupSRVTimeout(dnsTimeout, servicename, proto, host)
	if err != nil {
		return nil, err
	}
	if len(initialHosts) == 0 {
		return nil, balancer.ErrNoHosts
	}

	bal := &dnsSrvBalancer{
		serviceName: servicename,
		proto:       proto,
		host:        host,
		hosts:       initialHosts,
		interval:    updateInterval,
		counter:     0,
		quit:        make(chan int, 1),
		lock:        &sync.Mutex{},
		Timeout:     dnsTimeout,
	}

	// start update loop
	go bal.update()

	return bal, nil
}

func lookupSRVTimeout(timeout time.Duration, serviceName string, proto string, host string) ([]balancer.Host, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return lookupSRV(ctx, serviceName, proto, host)
}

func lookupSRV(ctx context.Context, servicename, proto, host string) ([]balancer.Host, error) {
	hosts := []balancer.Host{}

	_, addrs, err := net.DefaultResolver.LookupSRV(ctx, servicename, proto, host)
	if err != nil {
		return hosts, err
	}

	var firstErr error = nil

	for _, v := range addrs {
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, v.Target)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		for e := range ips {
			host := balancer.Host{
				Address: ips[e].IP,
				Port:    int(v.Port),
			}
			hosts = append(hosts, host)
		}
	}

	return hosts, firstErr
}

func (b *dnsSrvBalancer) Next() (balancer.Host, error) {
	// make sure to store a reference before we start
	hosts := b.hosts
	count := uint64(len(hosts))
	if count == 0 {
		return balancer.Host{}, balancer.ErrNoHosts
	}

	nextNum := atomic.AddUint64(&b.counter, 1)

	idx := nextNum % count

	return hosts[idx], nil
}

func (b *dnsSrvBalancer) update() {
	tick := time.NewTicker(b.interval)

	for {
		select {
		case <-tick.C:
			nextHostList, err := lookupSRVTimeout(b.Timeout, b.serviceName, b.proto, b.host)
			if err != nil {
				//  TODO: set hostList to empty?
				log.Printf("[SRVBalancer] error looking up dns='%v': %v", b.host, err)
			} else {
				if nextHostList != nil {
					log.Printf("[SRVBalancer] reloaded dns=%v hosts=%v", b.host, nextHostList)
					b.lock.Lock()
					b.hosts = nextHostList
					b.lock.Unlock()
				}
			}
		case <-b.quit:
			tick.Stop()
			return
		}
	}
}

func (b *dnsSrvBalancer) Close() error {
	// TODO: wait for exit
	b.quit <- 1

	return nil
}
