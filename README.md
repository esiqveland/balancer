# balancer

A simple client side load balancer for go applications.

`balancer` was made to provide easier access to DNS-based load balancing for go services running in kubernetes and was mainly built for http.Client.

## Scope

This project does not do health checking and does not monitor status of any hosts.
This is left up to decide for a consul DNS or kubernetes DNS, and assumes hosts returned are deemed healthy.

This library does not retry or otherwise try to fix problems, leaving this up to the caller.

`balancer` currently assumes that a lookup will return a non-empty set of initial hosts on startup.

## TODO

- [X] Any logging?
- [ ] Use option-func to configure net.Resolver
- [ ] Use option-func to configure debug mode (extra logging)
- [ ] Decide on a few error scenarios:
  - [X] `netbalancer`: implement a timeout?
  - [X] `netbalancer`: DNS lookup returned 0 hosts
  - [X] `netbalancer`: DNS lookup returned error. We just log the error, effectively ignoring it
    - [ ] on error, return lookup error?

## Known limitations

- No health checking of hosts.
- `netbalancer` does not respect TTL of dns records as this is not exposed by the Go code.
