//go:build linux

package os

import (
	"errors"
	"net"
	"syscall"
	"testing"
)

func TestConnectionBroken(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "non network error", err: errors.New("failure"), want: true},
		{name: "network down", err: &net.OpError{Op: "read", Net: "udp", Err: syscall.ENETDOWN}, want: true},
		{name: "network unreachable", err: &net.OpError{Op: "read", Net: "udp", Err: syscall.ENETUNREACH}, want: true},
		{name: "bad file descriptor", err: &net.OpError{Op: "read", Net: "udp", Err: syscall.EBADF}, want: true},
		{name: "temporary network error", err: &net.OpError{Op: "read", Net: "udp", Err: syscall.ECONNREFUSED}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConnectionBroken(tt.err); got != tt.want {
				t.Errorf("connectionBroken(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestNewConnectionControlOnUDPConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP() error = %v", err)
	}
	defer conn.Close()

	raw, err := conn.SyscallConn()
	if err != nil {
		t.Fatalf("SyscallConn() error = %v", err)
	}
	control := NewConnectionControl(net.Interface{})
	if err := control("udp", conn.LocalAddr().String(), raw); err != nil {
		t.Fatalf("connection control failed for an unbound interface: %v", err)
	}
}
