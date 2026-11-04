package transport

import (
	"errors"
	"fmt"
	"testing"
)

func TestTransportErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "not connected", err: ErrNotConnected, want: "not connected"},
		{name: "unexpected packet connection", err: UnexpectedPacketConnTypeError{Got: "*net.TCPConn"}, want: "*net.UDPConn cast failed for *net.TCPConn"},
		{name: "invalid header size", err: InvalidWireHeaderSizeError{Got: 31, Want: wireHeaderSize}, want: "failed to decode header, used 31 bytes while 32 bytes was expected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}

	if !errors.Is(ErrNotConnected, ErrNotConnected) {
		t.Error("ErrNotConnected does not match itself with errors.Is")
	}
	var typed NotConnectedError
	if !errors.As(ErrNotConnected, &typed) {
		t.Error("ErrNotConnected cannot be inspected as NotConnectedError")
	}

	var unexpectedType UnexpectedPacketConnTypeError
	if !errors.As(error(UnexpectedPacketConnTypeError{Got: "*net.TCPConn"}), &unexpectedType) {
		t.Error("UnexpectedPacketConnTypeError cannot be inspected with errors.As")
	}
	var invalidHeader InvalidWireHeaderSizeError
	if !errors.As(error(InvalidWireHeaderSizeError{Got: 31, Want: wireHeaderSize}), &invalidHeader) {
		t.Error("InvalidWireHeaderSizeError cannot be inspected with errors.As")
	}
}

func TestContentLengthExceededErrorFormatsMessage(t *testing.T) {
	err := ContentLengthExceededError(70000)
	want := fmt.Sprintf("maximum content length exceeded, got 70000 while max is %d", maxContentLength)
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
