package httpbalancer

import (
	"net"
	"net/http"
	"strconv"

	"github.com/esiqveland/balancer"
	"github.com/rs/zerolog"
)

func Wrap(client *http.Client, balancer balancer.Balancer) *http.Client {
	rt := NewBalancedRoundTripper(balancer, client.Transport)
	client.Transport = rt

	return client
}

func NewBalancedRoundTripper(balancer balancer.Balancer, delegate http.RoundTripper) http.RoundTripper {
	return &balancedRoundTripper{
		Delegate: delegate,
		Balancer: balancer,
	}
}

type balancedRoundTripper struct {
	Delegate http.RoundTripper
	Balancer balancer.Balancer
}

func (rt *balancedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	log := zerolog.Ctx(ctx)

	host, err := rt.Balancer.Next()
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("selected host=%v", host.Address.String())
	//var reqCopy http.Request = *req

	selectedHost := net.JoinHostPort(host.Address.String(), strconv.Itoa(host.Port))

	// strictly speaking, a RoundTripper is not allowed to mutate the request,
	// except for reading and Closing the req.Body so this might have consequences I am not aware of.
	req.URL.Host = selectedHost

	return rt.Delegate.RoundTrip(req)
}

var (
	// make sure we implement http.RoundTripper
	_ http.RoundTripper = &balancedRoundTripper{}
)
