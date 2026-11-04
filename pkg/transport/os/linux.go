//go:build linux

package os

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func NewConnectionControl(iface net.Interface) func(network string, address string, raw syscall.RawConn) error {
	return func(network string, address string, raw syscall.RawConn) error {
		var socketErrs []error
		if err := raw.Control(func(fd uintptr) {
			if err := unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, iface.Name); err != nil {
				socketErrs = append(socketErrs, err)
			}
			if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_BROADCAST, 1); err != nil {
				socketErrs = append(socketErrs, err)
			}
		}); err != nil {
			socketErrs = append(socketErrs, err)
		}
		return errors.Join(socketErrs...)
	}
}

func ConnectionBroken(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) { // [1, 2]
		var errno syscall.Errno
		if errors.As(opErr.Err, &errno) { // [1, 2]
			switch errno {
			case syscall.ENETDOWN, syscall.ENETUNREACH:
				// ENETDOWN: Network down
				// ENETUNREACH: Network unreachable
				fmt.Printf("[WARNING] Socket is closed: Interface of netwerk is down (OS code: %v)\n", errno)
				// Breek de loop af, de socket herstelt niet uit zichzelf
				return true
			case syscall.EBADF:
				// EBADF: Bad File Descriptor
				fmt.Println("[WARNING] Socket is closed: Invalid file descriptor.")
				return true
			default:
				//temporary, might start working again later
				return false
			}
		}
	}
	return true
}
