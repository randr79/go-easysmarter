// Package tlv encodes and parses the TLV content carried in transport message bodies.
package tlv

import (
	"encoding/binary"
	"fmt"
)

const (
	tlvTypeEnd uint16 = 0xffff
	maxLen            = int(^uint16(0))
)

type InvalidPayloadError struct {
	format string
	params []any
}

// Error returns the details of an invalid TLV payload.
func (i InvalidPayloadError) Error() string {
	return fmt.Sprintf(i.format, i.params...)
}

// TLV is a type-length-value item.
type TLV struct {
	Type  uint16
	Value []byte
}

// Length returns the number of bytes in the TLV value.
func (t TLV) Length() int {
	return len(t.Value)
}

// Parse parses TLVs from a message body, including its required terminator.
func Parse(data []byte) ([]TLV, error) {
	tlvs := make([]TLV, 0)
	for len(data) >= 4 {
		typeID := binary.BigEndian.Uint16(data[:2])
		size := int(binary.BigEndian.Uint16(data[2:4]))

		if typeID == tlvTypeEnd {
			if size != 0 {
				return nil, InvalidPayloadError{
					format: "invalid terminator length got %d expected 0",
					params: []any{size},
				}
			}
			if len(data) != 4 {
				return nil, InvalidPayloadError{
					format: "unexpected %d bytes after terminator",
					params: []any{len(data) - 4},
				}
			}
			return tlvs, nil
		}

		if size > len(data)-4 {
			return nil, InvalidPayloadError{
				format: "payload incomplete, need %d bytes, got %d ",
				params: []any{size + 4, len(data)},
			}
		}
		tlvs = append(tlvs, TLV{
			Type:  typeID,
			Value: append([]byte(nil), data[4:4+size]...),
		})
		data = data[4+size:]
	}
	return nil, InvalidPayloadError{
		format: "payload incomplete, missing terminator (got %d trailing bytes)",
		params: []any{len(data)},
	}
}

// Encode serializes TLVs for a message body and appends the required terminator.
func Encode(items []TLV) ([]byte, error) {
	payload := make([]byte, 0)
	for _, tlv := range items {
		if tlv.Type == tlvTypeEnd {
			return nil, InvalidPayloadError{
				format: "TLV type %d is reserved for the terminator",
				params: []any{tlvTypeEnd},
			}
		}
		if tl := tlv.Length(); tl > maxLen {
			return nil, InvalidPayloadError{
				format: "maximum tlv size is %d, got %d",
				params: []any{maxLen, tl},
			}
		}
		payload = binary.BigEndian.AppendUint16(payload, tlv.Type)
		payload = binary.BigEndian.AppendUint16(payload, uint16(tlv.Length()))
		payload = append(payload, tlv.Value...)
	}
	payload = binary.BigEndian.AppendUint16(payload, tlvTypeEnd)
	payload = binary.BigEndian.AppendUint16(payload, 0)
	return payload, nil
}
