package netbalancer

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	opt := Resolver(net.DefaultResolver)
	require.NotNil(t, opt)
}

type mockLookup struct {
	sleeptime  time.Duration
	resultsSRV func() (cname string, addrs []*net.SRV, err error)
	resultsIP  func() ([]net.IPAddr, error)
}

func (m *mockLookup) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	return m.resultsSRV()
}

func (m *mockLookup) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return m.resultsIP()
}
