package transport

import (
	"encoding/binary"
	"errors"
	"testing"
)

func TestStateDecodeRejectsInvalidPackets(t *testing.T) {
	table, err := NewStateTable([]byte("state-test-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	keyState, _ := table.Get([6]byte{})
	if _, _, err := keyState.decode([]byte{1, 2, 3}); err == nil {
		t.Fatal("decode() with a short header succeeded, want an error")
	}

	tests := []struct {
		name   string
		mutate func(*wireHeader)
	}{
		{name: "wrong version", mutate: func(header *wireHeader) { header.Version = 2 }},
		{name: "invalid method", mutate: func(header *wireHeader) { header.Method = 5 }},
		{name: "length exceeds packet", mutate: func(header *wireHeader) { header.Length = wireHeaderSize + 1 }},
		{name: "nonzero flags", mutate: func(header *wireHeader) { header.Flags = 1 }},
		{name: "nonzero CRC", mutate: func(header *wireHeader) { header.CRC = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := wireHeader{Version: 1, Length: wireHeaderSize}
			tt.mutate(&header)
			packet := encodeTestWireHeader(t, keyState, header)
			got, _, err := keyState.decode(packet)
			if err != nil {
				t.Fatalf("decode() error = %v, want nil for invalid-but-decodable header", err)
			}
			if got != nil {
				t.Errorf("decode() header = %#v, want nil", got)
			}
		})
	}
}

func TestStateEncodeRejectsOversizedMessage(t *testing.T) {
	table, err := NewStateTable([]byte("state-test-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	keyState, _ := table.Get([6]byte{})
	if _, err := keyState.encode(wireHeader{}, make([]byte, maxContentLength+1)); err == nil {
		t.Fatal("encode() succeeded for oversized message, want an error")
	} else {
		var payloadErr ContentLengthExceededError
		if !errors.As(err, &payloadErr) {
			t.Errorf("encode() error type = %T, want ContentLengthExceededError", err)
		}
	}
}

func TestStateEncodeDecodeRoundTrip(t *testing.T) {
	table, err := NewStateTable([]byte("state-test-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	keyState, _ := table.Get([6]byte{})
	header := wireHeader{
		Version:   1,
		Method:    byte(MethodDiscover),
		SwitchMac: [6]byte{1, 2, 3, 4, 5, 6},
		RequestId: 42,
	}
	body := []byte{0xff, 0xff, 0, 0}
	packet, err := keyState.encode(header, body)
	if err != nil {
		t.Fatalf("encode() error = %v", err)
	}
	gotHeader, gotBody, err := keyState.decode(packet)
	if err != nil {
		t.Fatalf("decode() error = %v", err)
	}
	if gotHeader == nil || gotHeader.RequestId != header.RequestId || int(gotHeader.Length) != wireHeaderSize+len(body) {
		t.Errorf("decoded header = %#v, want request %d and length %d", gotHeader, header.RequestId, wireHeaderSize+len(body))
	}
	if string(gotBody) != string(body) {
		t.Errorf("decoded body = %x, want %x", gotBody, body)
	}
}

func encodeTestWireHeader(t *testing.T, keyState *State, header wireHeader) []byte {
	t.Helper()
	raw, err := binary.Append(nil, binary.BigEndian, header)
	if err != nil {
		t.Fatalf("binary.Append() error = %v", err)
	}
	return keyState.Encryption.XORKey(raw)
}
