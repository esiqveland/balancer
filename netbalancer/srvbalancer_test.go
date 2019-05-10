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

