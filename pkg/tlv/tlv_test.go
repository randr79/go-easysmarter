package tlv

import (
	"encoding/binary"
	"errors"
	"testing"
)

func TestTLVEncodeAndParseRoundTrip(t *testing.T) {
	want := []TLV{
		{Type: 1, Value: []byte("first")},
		{Type: 0x1234, Value: []byte{0, 1, 2, 255}},
	}
	encoded, err := Encode(want)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	wantEncoded := []byte{
		0, 1, 0, 5, 'f', 'i', 'r', 's', 't',
		0x12, 0x34, 0, 4, 0, 1, 2, 255,
		0xff, 0xff, 0, 0,
	}
	if string(encoded) != string(wantEncoded) {
		t.Fatalf("Encode() = %x, want %x", encoded, wantEncoded)
	}
	got, err := Parse(encoded)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("Parse() returned %d TLVs, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Type != want[i].Type || string(got[i].Value) != string(want[i].Value) {
			t.Errorf("TLV %d = %#v, want %#v", i, got[i], want[i])
		}
	}
	if got := (TLV{Value: []byte{1, 2, 3}}).Length(); got != 3 {
		t.Errorf("Length() = %d, want 3", got)
	}
}

func TestParseTLVRejectsMalformedPayloads(t *testing.T) {
	incompleteValue := []byte{0, 1, 0, 2, 0}
	invalidTerminator := []byte{0, 1, 0, 0}
	invalidTerminatorLength := []byte{0xff, 0xff, 0, 1}
	trailingData := []byte{0xff, 0xff, 0, 0, 0, 1, 0, 0}
	tests := []struct {
		name string
		data []byte
	}{
		{name: "short header", data: []byte{1, 2, 3}},
		{name: "incomplete value", data: incompleteValue},
		{name: "invalid terminator", data: invalidTerminator},
		{name: "invalid terminator length", data: invalidTerminatorLength},
		{name: "trailing data", data: trailingData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.data); err == nil {
				t.Fatal("Parse() succeeded, want an error")
			} else {
				var payloadErr InvalidPayloadError
				if !errors.As(err, &payloadErr) {
					t.Errorf("Parse() error type = %T, want InvalidPayloadError", err)
				}
			}
		})
	}
}

func TestEncodeTLVRejectsOversizedValue(t *testing.T) {
	value := make([]byte, maxLen+1)
	if _, err := Encode([]TLV{{Type: 1, Value: value}}); err == nil {
		t.Fatal("Encode() succeeded for an oversized value, want an error")
	} else {
		var payloadErr InvalidPayloadError
		if !errors.As(err, &payloadErr) {
			t.Errorf("Encode() error type = %T, want InvalidPayloadError", err)
		}
	}
}

func TestEncodeTLVRejectsTerminatorType(t *testing.T) {
	if _, err := Encode([]TLV{{Type: tlvTypeEnd}}); err == nil {
		t.Fatal("Encode() succeeded for the reserved terminator type, want an error")
	}
}

func TestInvalidPayloadErrorFormatsMessage(t *testing.T) {
	err := InvalidPayloadError{format: "size %d exceeds %d", params: []any{10, 5}}
	if got, want := err.Error(), "size 10 exceeds 5"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestTLVTerminatorEncoding(t *testing.T) {
	encoded, err := Encode(nil)
	if err != nil {
		t.Fatalf("Encode(nil) error = %v", err)
	}
	if len(encoded) != 4 || binary.BigEndian.Uint16(encoded[:2]) != tlvTypeEnd || binary.BigEndian.Uint16(encoded[2:]) != 0 {
		t.Errorf("empty TLV encoding = %x, want ffff0000", encoded)
	}
}
