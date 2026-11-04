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

type MACAddress [6]byte

func (m MACAddress) String() string {
	const hexDigit = "0123456789abcdef"
	buf := make([]byte, 0, 17)
	for i, b := range m {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, hexDigit[b>>4], hexDigit[b&0xF])
	}
	return string(buf)
}

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

//go:generate stringer -type=Method -trimprefix=Method
type Method byte

const (
	MethodDiscover     Method = 0
	MethodGetRequest   Method = 1
	MethodGetResponse  Method = 2
	MethodPostRequest  Method = 3
	MethodPostResponse Method = 4
)

type Message struct {
	Method    Method
	SwitchMac MACAddress
	ClientMac MACAddress

	RequestId uint16
	Fragment  uint16 // 2 bytes
	State     StatusCode
	Body      []byte
}

const maxContentLength int = int(^uint16(0)) - 32
