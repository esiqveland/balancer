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

type srvOption func(*dnsSrvBalancer)

func Resolver(r *net.Resolver) srvOption {
	return func(d *dnsSrvBalancer) {
		d.resolver = r
	}
}

func UpdateInterval(d time.Duration) srvOption {
	return func(dsb *dnsSrvBalancer) {
		dsb.interval = d
	}
}

func Timeout(d time.Duration) srvOption {
	return func(dsb *dnsSrvBalancer) {
		dsb.timeout = d
	}
}

// dnsSrvBalancer holds all the pieces for an updating SRV lookup.
type dnsSrvBalancer struct {
	resolver    *net.Resolver
	timeout     time.Duration
	serviceName string
	proto       string
	host        string
	hosts       []balancer.Host
	counter     uint64
	interval    time.Duration
	quit        chan int
	lock        *sync.Mutex
}

// NewSRV returns a Balancer that uses dns lookups from net.LookupSRV to reload a set of hosts every updateInterval.
// We can not use TTL from dns because TTL is not exposed by the Go stdlib.
func NewSRV(servicename, proto, host string, opts ...srvOption) (balancer.Balancer, error) {
	bal := &dnsSrvBalancer{
		serviceName: servicename,
		resolver:    net.DefaultResolver,
		proto:       proto,
		host:        host,
		hosts:       nil,
		counter:     0,
		quit:        make(chan int, 1),
		lock:        &sync.Mutex{},
		interval:    time.Second * 5,
		timeout:     time.Second * 2,
	}

	for _, opt := range opts {
		opt(bal)
	}

	initialHosts, err := lookupSRVTimeout(bal.timeout, bal.resolver, servicename, proto, host)
	if err != nil {
		return nil, err
	}
	if len(initialHosts) == 0 {
		return nil, balancer.ErrNoHosts
	}
	bal.hosts = initialHosts

	// start update loop
	go bal.loop()

	return bal, nil
}

func lookupSRVTimeout(timeout time.Duration, resolver *net.Resolver, serviceName string, proto string, host string) ([]balancer.Host, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return lookupSRV(ctx, resolver, serviceName, proto, host)
}

func lookupSRV(ctx context.Context, resolver *net.Resolver, servicename, proto, host string) ([]balancer.Host, error) {
	hosts := []balancer.Host{}

	_, addrs, err := resolver.LookupSRV(ctx, servicename, proto, host)
	if err != nil {
		return hosts, err
	}

	var firstErr error = nil

	for _, v := range addrs {
		ips, err := resolver.LookupIPAddr(ctx, v.Target)
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

func (b *dnsSrvBalancer) update() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	nextHostList, err := lookupSRVTimeout(b.timeout, b.resolver, b.serviceName, b.proto, b.host)
	if err != nil {
		//  TODO: set hostList to empty?
		log.Printf("[SRVBalancer] error looking up dns='%v': %v", b.host, err)
	} else {
		if nextHostList != nil {
			log.Printf("[SRVBalancer] reloaded dns=%v hosts=%v", b.host, nextHostList)
			b.hosts = nextHostList
		}
	}
	return err
}

func (b *dnsSrvBalancer) loop() {
	tick := time.NewTicker(b.interval)

	for {
		select {
		case <-tick.C:
			b.update()
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

var (
	// make sure we implement Balancer interface
	_ balancer.Balancer = &dnsSrvBalancer{}
)
