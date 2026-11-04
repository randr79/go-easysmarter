package transport

import (
	"context"
	"fmt"
	"net"

	transportOs "github.com/randr79/go-easysmarter/pkg/transport/os"
)

// UDPConn is the subset of *net.UDPConn used by the transport, allowing test fakes.
type UDPConn interface {
	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	WriteToUDP([]byte, *net.UDPAddr) (int, error)
	Close() error
}

// UDPConnFactory opens the UDP connection for the given port and interface.
type UDPConnFactory func(port uint16, iface net.Interface) (UDPConn, error)

func NewUDPConn(port uint16, iface net.Interface) (UDPConn, error) {
	lc := net.ListenConfig{Control: transportOs.NewConnectionControl(iface)}
	conn, err := lc.ListenPacket(context.Background(), "udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}
	udpconn, ok := conn.(*net.UDPConn)
	if !ok {
		_ = conn.Close()
		return nil, UnexpectedPacketConnTypeError{Got: fmt.Sprintf("%T", conn)}
	}
	return udpconn, nil
}
