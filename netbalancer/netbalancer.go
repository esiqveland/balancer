package netbalancer

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/esiqveland/balancer"
	"github.com/pkg/errors"
)

// NewNetBalancer returns a Balancer that uses dns lookups from net.Lookup* to reload a set of hosts every updateInterval.
// We can not use TTL from dns because TTL is not exposed by the Go calls.
func New(host string, port int, updateInterval, timeout time.Duration) (balancer.Balancer, error) {
	initialHosts, err := lookupTimeout(timeout, host, port)
	if err != nil {
		return nil, err
	}
	if len(initialHosts) == 0 {
		return nil, balancer.ErrNoHosts
	}

	bal := &dnsBalancer{
		lookupAddress: host,
		port:          port,
		hosts:         initialHosts,
		interval:      updateInterval,
		counter:       0,
		quit:          make(chan int, 1),
		lock:          &sync.Mutex{},
		Timeout:       timeout,
	}

	// start update loop
	go bal.update()

	return bal, nil
}

type dnsBalancer struct {
	lookupAddress string
	port          int
	hosts         []balancer.Host
	counter       uint64
	interval      time.Duration
	quit          chan int
	lock          *sync.Mutex
	Timeout       time.Duration
}

func (b *dnsBalancer) Next() (balancer.Host, error) {
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

func (b *dnsBalancer) update() {
	tick := time.NewTicker(b.interval)

	for {
		select {
		case <-tick.C:
			nextHostList, err := lookupTimeout(b.Timeout, b.lookupAddress, b.port)
			if err != nil {
				//  TODO: set hostList to empty?
				log.Printf("[SrvBalancer] error looking up dns='%v': %v", b.lookupAddress, err)
			} else {
				if nextHostList != nil {
					log.Printf("[DnsBalancer] reloaded dns=%v hosts=%v", b.lookupAddress, nextHostList)
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

func lookupTimeout(timeout time.Duration, host string, port int) ([]balancer.Host, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return lookup(ctx, host, port)
}

func lookup(ctx context.Context, host string, port int) ([]balancer.Host, error) {
	hosts := []balancer.Host{}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return hosts, errors.Wrapf(err, "Error looking up host=%v", host)
	}

	for k := range ips {
		entry := balancer.Host{
			Address: ips[k].IP,
			Port:    port,
		}
		hosts = append(hosts, entry)
	}

	return hosts, nil
}

func (b *dnsBalancer) Close() error {
	// TODO: wait for exit
	b.quit <- 1

	return nil
}
