# balancer

A simple client side load balancer for go applications.

`balancer` was made to provide easier access to DNS-based load balancing for go services running in kubernetes and was mainly built for http.Client.

Grpc has its own DNS load balancer, use that one.

## Scope

This project does not do health checking and does not monitor status of any hosts.
This is left up to decide for a consul DNS or kubernetes DNS, and assumes hosts returned are deemed healthy.

This library does not retry or otherwise try to fix problems, leaving this up to the caller.

`balancer` currently assumes that a lookup will return a non-empty set of initial hosts on startup.

## TODO

- [X] Any logging?
- [X] Decide on a few error scenarios:
  - [X] `netbalancer`: implement a timeout?
  - [X] `netbalancer`: DNS lookup returned 0 hosts
  - [X] `netbalancer`: DNS lookup returned error. We just log the error, effectively ignoring it
    - [X] on error, return lookup error?
    - [X] setting the hostlist to empty on error seems a bit drastic
      - [X] on lookup error, we keep the old list and log error
- [ ] Tests
- [ ] Use option-func to configure debug mode (extra logging)

## Known limitations

- No health checking of hosts.
- `netbalancer` does not respect TTL of dns records as this is not exposed by the Go code.
