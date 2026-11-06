package transport

import (
	"errors"
	"fmt"
)

const (
	// not used here but can't think of a better place to put it
	PortSwitch  uint16 = 29808
	PortManager uint16 = 29809
)

type NotConnectedError struct{}

func (NotConnectedError) Error() string {
	return "not connected"
}

var ErrNotConnected error = NotConnectedError{}

type UnexpectedPacketConnTypeError struct {
	Got string
}

func (e UnexpectedPacketConnTypeError) Error() string {
	return fmt.Sprintf("*net.UDPConn cast failed for %s", e.Got)
}

type InvalidWireHeaderSizeError struct {
	Got  int
	Want int
}

func (e InvalidWireHeaderSizeError) Error() string {
	return fmt.Sprintf("failed to decode header, used %d bytes while %d bytes was expected", e.Got, e.Want)
}

var ErrNoConnFactory = errors.New("transport: connection factory not set")

type ContentLengthExceededError int

func (c ContentLengthExceededError) Error() string {
	return fmt.Sprintf("maximum content length exceeded, got %d while max is %d", c, maxContentLength)
}

type StatusCode uint32

const (
	StatusCodeSuccess StatusCode = 0

	GetStatusCodeUnsupported StatusCode = 65535 - 2

	SetStatusCodeRsaSuccess           StatusCode = 4098 //0x1002
	SetStatusCodeInvalidRequest       StatusCode = 3
	SetStatusCodeIncorrectOldPassword StatusCode = 5
	SetStatusCodeInvalidCredentials   StatusCode = 7
)

type Header struct {
	Method    Method
	SwitchMac MACAddress
	ClientMac MACAddress

	RequestId uint16
	Fragment  uint16 // 2 bytes
	State     StatusCode
}

type Message struct {
	Header
	Body []byte
}

const maxContentLength int = int(^uint16(0)) - 32
