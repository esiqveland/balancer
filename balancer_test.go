package balancer

import (
	"fmt"
	"net"
	"testing"
)

func TestStringer(t *testing.T) {
	expected := "192.168.0.1:123"
	h := Host{Address: net.IPv4(192, 168, 0, 1), Port: 123}
	if h.String() != expected {
		t.Fatalf("wanted=%v but got %v", expected, h.String())
	}
}

func TestPrint(t *testing.T) {
	expected := "192.168.0.1:123"
	h := Host{Address: net.IPv4(192, 168, 0, 1), Port: 123}

	val1 := fmt.Sprintf("%v", h)
	if val1 != expected {
		t.Fatalf("wanted=%v but got %v", expected, val1)
	}

	h2 := &Host{Address: net.IPv4(192, 168, 0, 1), Port: 123}
	val2 := fmt.Sprintf("%v", h2)
	if val2 != expected {
		t.Fatalf("wanted=%v but got %v", expected, val2)
	}
}
