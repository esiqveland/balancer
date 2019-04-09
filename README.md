# balancer

A simple client side load balancer for go applications.

`balancer` was made to provide easier access to DNS-based load balancing for go services running in kubernetes and was mainly built for http.Client.

## Scope

This project does not do health checking and does not monitor status of any hosts.
This is left up to decide for a consul DNS or kubernetes DNS, and assumes hosts returned are deemed healthy.

`balancer` currently assumes that a lookup will return a non-empty set of initial hosts on startup.

## TODO

- [ ] Decide on a few error scenarios:
  - [ ] DNS lookup hangs/times out
  - [ ] DNS lookup returned 0 hosts

## Known limitations

- No health checking of hosts.
- Does not respect TTL of dns records as this is not exposed by the Go code.
